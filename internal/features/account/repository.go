// internal/features/account/repository.go
package account

import (
	"context"
	"errors"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"

	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

// Repository persists account profiles. Lookups return (nil, nil) when absent;
// updates and deletions return ErrUserNotFound when the target no longer exists.
// Infrastructure failures wrap repository.ErrRepositoryFailure and their cause.
type Repository interface {
	// CRUD basico
	Create(ctx context.Context, u *user.User) error
	Update(ctx context.Context, u *user.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Buscas por atributos do usuario

	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByCPF(ctx context.Context, cpf string) (*user.User, error)
	FindByAuthIdentity(ctx context.Context, issuer string, subject string) (*user.User, error)
}
