// internal/domain/user_repository.go
package repositories

import (
	"context"

	"sonnda-api/internal/core/domain/identity"
)

type UserRepository interface {
	FindByAuthIdentity(ctx context.Context, provider, subject string) (*identity.User, error)
	FindByEmail(ctx context.Context, email string) (*identity.User, error)
	FindByID(ctx context.Context, id string) (*identity.User, error)
	UpdateAuthIdentity(ctx context.Context, id string, provider, subject string) (*identity.User, error)
	Create(ctx context.Context, u *identity.User) error
}
