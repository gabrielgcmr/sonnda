// internal/domain/user_repository.go
package repositories

import (
	"context"
	"sonnda-api/internal/core/domain"
)

type UserRepository interface {
	FindByAuthIdentity(ctx context.Context, provider, subject string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	UpdateAuthIdentity(ctx context.Context, id, provider, subject string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
}
