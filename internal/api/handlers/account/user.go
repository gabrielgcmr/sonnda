// internal/api/handlers/account/user.go
package account

import (
	"context"

	base "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	usersvc "github.com/gabrielgcmr/sonnda/internal/application/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/registration"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/google/uuid"
)

type UserService interface {
	Update(ctx context.Context, input usersvc.UserUpdateInput) (*user.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	ListMyPatients(ctx context.Context, userID uuid.UUID, limit, offset int) (*usersvc.MyPatientsOutput, error)
}

type UserHandler = base.UserHandler

func NewUserHandler(regUC registrationuc.UseCase, userSvc UserService) *UserHandler {
	return base.NewUserHandler(regUC, userSvc)
}
