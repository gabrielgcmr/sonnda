// internal/adapters/inbound/http/middleware/auth.go
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/app/apperr"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/identity"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
)

// Cookie de sessão do browser (seu nome atual).
const SessionCookieName = "__session"

// Erros “baixos” (infra) para facilitar mapear em AppError.
var (
	ErrSessionCookieMissing       = errors.New("session cookie missing")
	ErrAuthorizationHeaderMissing = errors.New("authorization header missing")
	ErrAuthorizationHeaderInvalid = errors.New("invalid authorization header format")
)

// Core = só resolve identidade.
// Não conhece gin.
// Não escreve resposta.
// Não aborta request.
type Core struct {
	identityService ports.IdentityService
}

func NewCore(identityService ports.IdentityService) *Core {
	return &Core{identityService: identityService}
}

// AuthenticateFromSessionCookie é o fluxo típico do WEB (app.*):
// - lê cookie __session
// - valida no provider (ex: Firebase/Supabase/qualquer)
// - retorna Identity ou AppError
func (c *Core) AuthenticateFromSessionCookie(ctx context.Context, r *http.Request) (*identity.Identity, *apperr.AppError) {
	cookie, err := extractSessionCookie(r)
	if err != nil {
		return nil, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
			Cause:   err,
		}
	}

	id, verr := c.identityService.VerifySessionCookie(ctx, cookie)
	if verr != nil {
		return nil, &apperr.AppError{
			Code:    apperr.AUTH_TOKEN_INVALID,
			Message: "sessão inválida ou expirada",
			Cause:   verr,
		}
	}
	return id, nil
}

// AuthenticateFromBearerToken é o fluxo típico da API (api.*):
// - lê Authorization: Bearer <token>
// - valida no provider
func (c *Core) AuthenticateFromBearerToken(ctx context.Context, r *http.Request) (*identity.Identity, *apperr.AppError) {
	token, err := extractBearerToken(r)
	if err != nil {
		// Se nem veio header, é “AUTH_REQUIRED”.
		if errors.Is(err, ErrAuthorizationHeaderMissing) {
			return nil, &apperr.AppError{
				Code:    apperr.AUTH_REQUIRED,
				Message: "autenticação necessária",
				Cause:   err,
			}
		}
		// Se veio mas está malformado, é inválido.
		return nil, &apperr.AppError{
			Code:    apperr.AUTH_TOKEN_INVALID,
			Message: "token inválido",
			Cause:   err,
		}
	}

	id, verr := c.identityService.VerifyToken(ctx, token)
	if verr != nil {
		return nil, &apperr.AppError{
			Code:    apperr.AUTH_TOKEN_INVALID,
			Message: "token inválido ou expirado",
			Cause:   verr,
		}
	}
	return id, nil
}

// Se você tiver endpoints “mistos” (aceita cookie OU bearer no mesmo host),
// use esse método. Se for separar por subdomínio (app vs api), você provavelmente
// nem precisa dele.
func (c *Core) AuthenticateSessionThenBearer(ctx context.Context, r *http.Request) (*identity.Identity, *apperr.AppError) {
	if id, err := c.AuthenticateFromSessionCookie(ctx, r); err == nil {
		return id, nil
	}
	return c.AuthenticateFromBearerToken(ctx, r)
}

func extractSessionCookie(r *http.Request) (string, error) {
	ck, err := r.Cookie(SessionCookieName)
	if err != nil || ck == nil || ck.Value == "" {
		return "", ErrSessionCookieMissing
	}
	return ck.Value, nil
}

func extractBearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", ErrAuthorizationHeaderMissing
	}
	if !strings.HasPrefix(h, "Bearer ") {
		return "", ErrAuthorizationHeaderInvalid
	}
	return strings.TrimPrefix(h, "Bearer "), nil
}
