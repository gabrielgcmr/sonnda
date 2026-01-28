// internal/adapters/inbound/http/shared/auth/core.go
package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

const (
	DefaultSessionCookieName = "__session"
)

// Core v2: orquestra autenticação HTTP -> IdentityProvider.
// - Não faz "gambiarra": apenas extrai credencial e delega validação ao provider.
// - Centraliza o mapeamento de AuthStatus -> AppError.
type Core struct {
	provider          security.IdentityProvider
	sessionCookieName string

	// Opcional: timeouts curtos para auth (evita request travar em JWKS/Redis).
	bearerTimeout time.Duration
	cookieTimeout time.Duration
}

type CoreOption func(*Core)

func WithSessionCookieName(name string) CoreOption {
	return func(c *Core) {
		if strings.TrimSpace(name) != "" {
			c.sessionCookieName = name
		}
	}
}

func WithTimeouts(bearer, cookie time.Duration) CoreOption {
	return func(c *Core) {
		if bearer > 0 {
			c.bearerTimeout = bearer
		}
		if cookie > 0 {
			c.cookieTimeout = cookie
		}
	}
}

func NewCore(provider security.IdentityProvider, opts ...CoreOption) *Core {
	c := &Core{
		provider:          provider,
		sessionCookieName: DefaultSessionCookieName,
		bearerTimeout:     2 * time.Second,
		cookieTimeout:     2 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AuthenticateFromBearerToken valida token e retorna Identity.
// - token ausente/ inválido => Unauthorized(AUTH_REQUIRED)
// - erro técnico (JWKS, rede, bug) => Internal(INTERNAL_ERROR)
func (c *Core) AuthenticateFromBearerToken(ctx context.Context, bearerToken string) (*security.Identity, error) {
	if strings.TrimSpace(bearerToken) == "" {
		return nil, apperr.Unauthorized("missing bearer token")
	}

	ctxto, cancel := context.WithTimeout(ctx, c.bearerTimeout)
	defer cancel()

	res, err := c.provider.AuthenticateBearerToken(ctxto, bearerToken)
	if err != nil {
		// erro técnico (JWKS down, rede, bug) -> 500
		return nil, apperr.Internal("falha ao autenticar token", err)
	}

	switch res.Status {
	case security.Authenticated:
		return res.Identity, nil
	case security.Unauthenticated:
		// token inválido/expirado -> 401
		return nil, apperr.Unauthorized("token inválido ou expirado")
	case security.Error:
		return nil, apperr.Internal("falha ao autenticar token", nil)
	default:
		// fallback defensivo
		return nil, apperr.Internal("status de autenticação inesperado", nil)
	}
}

// AuthenticateFromSessionCookie valida cookie __session e retorna Identity.
//   - Authenticated => retorna Identity
//   - Unauthenticated => Unauthorized (sessão expirada/inválida)
//   - Erro técnico (Redis down) => Internal
func (c *Core) AuthenticateFromSessionCookie(ctx context.Context, r *http.Request) (*security.Identity, error) {
	// Tenta ler o cookie; se não existe, Cookie() retorna ErrNoCookie
	//1. Extrai o ID de sessão do cookie nomeado (ex: __session)
	cookie, err := r.Cookie(c.sessionCookieName)
	if err != nil || cookie == nil {
		return nil, apperr.Unauthorized("cookie de sessão ausente")
	}
	//2. Extrai o valor do cookie (ID de sessão)
	sid := strings.TrimSpace(cookie.Value)
	if sid == "" {
		return nil, apperr.Unauthorized("cookie de sessão ausente")
	}
	//3. Valida o ID de sessão via IdentityProvider
	ctxto, cancel := context.WithTimeout(ctx, c.cookieTimeout)
	defer cancel()

	//4. Processa o resultado
	res, err := c.provider.AuthenticateCookie(ctxto, sid)
	if err != nil {
		return nil, apperr.Internal("falha ao autenticar sessão", err)
	}

	//5. Mapeia status para retorno
	switch res.Status {
	case security.Authenticated:
		return res.Identity, nil
	case security.Unauthenticated:
		// sessão não existe / expirou / inválida
		return nil, apperr.Unauthorized("Sessão não existe, inválida ou expirada")
	case security.Error:
		return nil, apperr.Internal("erro ao autenticar sessão", nil)
	default:
		return nil, apperr.Internal("status de autenticação inesperado", nil)
	}
}
