package usersvc

import (
	"context"

	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, input UserCreateInput) (*user.User, error)
	Update(ctx context.Context, input UserUpdateInput) (*user.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error

	// SoftDelete performs a logical delete.
	//
	// Semantics:
	// - If the user does not exist, returns NOT_FOUND.
	// - If the user exists, repeated calls are idempotent (calling SoftDelete again is considered success).
	//
	// This is intentional so callers can safely retry without turning a previously successful delete into an error.
	SoftDelete(ctx context.Context, userID uuid.UUID) error
}
