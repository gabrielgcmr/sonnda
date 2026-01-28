package registration

import (
	"context"
	"errors"

	professionalsvc "github.com/gabrielgcmr/sonnda/internal/app/services/professional"
	usersvc "github.com/gabrielgcmr/sonnda/internal/app/services/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	auth "github.com/gabrielgcmr/sonnda/internal/domain/ports/auth"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/data"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type UseCase interface {
	Register(ctx context.Context, input RegisterInput) (*user.User, error)
}

type usecase struct {
	userRepo data.UserRepo
	userSvc  usersvc.Service
	profSvc  professionalsvc.Service
	authSvc  auth.IdentityProvider
}

var _ UseCase = (*usecase)(nil)

func New(userRepo data.UserRepo, userSvc usersvc.Service, profSvc professionalsvc.Service, authSvc auth.IdentityProvider) *usecase {
	return &usecase{
		userRepo: userRepo,
		userSvc:  userSvc,
		profSvc:  profSvc,
		authSvc:  authSvc,
	}
}

func (u *usecase) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	// Verificar se usuǭrio jǭ existe
	existing, err := u.userRepo.FindByAuthIdentity(ctx, input.Provider, input.Subject)
	if err != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha ao verificar registro",
			Cause:   err,
		}
	}
	if existing != nil {
		return nil, &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "usuǭrio jǭ cadastrado",
		}
	}

	createdUser, err := u.userSvc.Create(ctx, usersvc.UserCreateInput{
		Provider:    input.Provider,
		Subject:     input.Subject,
		Email:       input.Email,
		AccountType: input.AccountType,
		FullName:    input.FullName,
		BirthDate:   input.BirthDate,
		CPF:         input.CPF,
		Phone:       input.Phone,
	})
	if err != nil {
		return nil, err
	}

	if input.AccountType != user.AccountTypeProfessional {
		return createdUser, nil
	}

	if input.Professional == nil {
		return nil, apperr.Validation("dados do profissional invǭlidos")
	}

	_, err = u.profSvc.Create(ctx, professionalsvc.CreateInput{
		UserID:             createdUser.ID,
		Kind:               input.Professional.Kind,
		RegistrationNumber: input.Professional.RegistrationNumber,
		RegistrationIssuer: input.Professional.RegistrationIssuer,
		RegistrationState:  input.Professional.RegistrationState,
	})
	if err == nil {
		return createdUser, nil
	}

	rollbackErr := u.rollbackCreatedUser(ctx, createdUser)
	if rollbackErr != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   errors.Join(err, rollbackErr),
		}
	}

	return nil, err
}

func (u *usecase) rollbackCreatedUser(ctx context.Context, createdUser *user.User) error {
	if createdUser == nil {
		return nil
	}

	var rollbackErr error

	if createdUser.AuthIssuer == "firebase" && u.authSvc != nil {
		rollbackErr = errors.Join(rollbackErr, u.authSvc.DisableUser(ctx, createdUser.AuthSubject))
	}

	return rollbackErr
}
