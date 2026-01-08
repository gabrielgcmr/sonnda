package registrationsvc

import (
	"context"

	"sonnda-api/internal/domain/model/user"
)

type Service interface {
	Register(ctx context.Context, input RegisterInput) (*user.User, error)
}
