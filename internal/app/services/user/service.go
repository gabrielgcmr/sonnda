package usersvc

import (
	"context"

	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type UserService interface {
	Register(ctx context.Context, input UserRegisterInput) (*user.User, error)
	Update(ctx context.Context, input UserUpdateInput) (*user.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}
