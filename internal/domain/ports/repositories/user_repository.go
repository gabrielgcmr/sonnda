// internal/domain/user_repository.go
package repositories

import (
	"context"

	"sonnda-api/internal/domain/entities/user"

	"github.com/google/uuid"
)

type UserRepository interface {
	// CRUD basico
	Save(ctx context.Context, u *user.User) error
	Update(ctx context.Context, u *user.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Buscas por atributos do usuario
	FindByAuthIdentity(ctx context.Context, provider, subject string) (*user.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByCPF(ctx context.Context, cpf string) (*user.User, error)
}
