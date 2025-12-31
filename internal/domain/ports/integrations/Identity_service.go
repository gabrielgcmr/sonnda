// internal/domain/ports/services/Identity_service.go
package integrations

import (
	"context"
	"sonnda-api/internal/domain/entities/identity"
)

type IdentityService interface {
	ProviderName() string
	VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error)

	DisableUser(ctx context.Context, subject string) error
}
