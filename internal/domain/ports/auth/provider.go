// internal/domain/ports/integration/Identity_service.go
package auth

import (
	"context"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/identity"
)

type IdentityProvider interface {
	ProviderName() string
	VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error)
	VerifySessionCookie(ctx context.Context, sessionCookie string) (*identity.Identity, error)

	CreateSessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error)
	RevokeSessions(ctx context.Context, subject string) error

	DisableUser(ctx context.Context, subject string) error
}
