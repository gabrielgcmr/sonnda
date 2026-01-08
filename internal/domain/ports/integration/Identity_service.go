// internal/domain/ports/integrations/Identity_service.go
package integration

import (
	"context"
	"sonnda-api/internal/domain/model/identity"
)

type IdentityService interface {
	ProviderName() string
	VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error)

	DisableUser(ctx context.Context, subject string) error
}
