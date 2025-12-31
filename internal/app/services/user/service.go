package usersvc

import (
	"context"

	"sonnda-api/internal/domain/entities/user"

	"github.com/google/uuid"
)

type Service interface {
	Register(ctx context.Context, input RegisterInput) (*user.User, error)
	Update(ctx context.Context, input UpdateInput) (*user.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}
