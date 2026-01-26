// internal/domain/user_repository.go
package repository

import (
	"context"

	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type User interface {
	// CRUD basico
	Create(ctx context.Context, u *user.User) error
	Update(ctx context.Context, u *user.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Buscas por atributos do usuario

	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByCPF(ctx context.Context, cpf string) (*user.User, error)
	FindByAuthIdentity(ctx context.Context, provider, subject string) (*user.User, error)
}
