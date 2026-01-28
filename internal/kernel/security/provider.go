// internal/shared/security/provider.go
package security

import (
	"context"
	"time"
)

type AuthMethod string

const (
	AuthMethodBearer AuthMethod = "bearer"
	AuthMethodCookie AuthMethod = "cookie"
)

type AuthStatus string

const (
	// O usuário provou quem é com sucesso.
	Authenticated AuthStatus = "authenticated"

	// O usuário não enviou credenciais ou elas são inválidas,
	// mas o sistema funcionou corretamente.
	Unauthenticated AuthStatus = "unauthenticated"
	// Ocorreu um erro técnico (banco fora, erro de parsing, etc).
	// O usuário deve receber um 500 Internal Server Error, não um 401.
	Error AuthStatus = "error"
)

type IdentityProvider interface {
	Name() string

	AuthenticateBearerToken(ctx context.Context, bearerToken string) (*AuthResult, error)
	AuthenticateCookie(ctx context.Context, cookieValue string) (*AuthResult, error)
}

type AuthResult struct {
	Status    AuthStatus
	Identity  *Identity
	Method    *AuthMethod
	ExpiresAt *time.Time
	Claims    map[string]any
}

// Dica Extra: Helpers para limpar seu código nos Handlers
func (s AuthStatus) IsAuthenticated() bool {
	return s == Authenticated
}
