package registration

import (
	"context"
	"errors"

	"sonnda-api/internal/app/apperr"
	professionalsvc "sonnda-api/internal/app/services/professional"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/model/user"
	auth "sonnda-api/internal/domain/ports/auth"
	"sonnda-api/internal/domain/ports/data"
)

type UseCase interface {
	Register(ctx context.Context, input RegisterInput) (*user.User, error)
}

type usecase struct {
	userRepo data.UserRepo
	userSvc  usersvc.Service
	profSvc  professionalsvc.Service
	authSvc  auth.IdentityService
}

var _ UseCase = (*usecase)(nil)

func New(userRepo data.UserRepo, userSvc usersvc.Service, profSvc professionalsvc.Service, authSvc auth.IdentityService) *usecase {
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
			Message: "falha tǸcnica",
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

	if createdUser.AuthProvider == "firebase" && u.authSvc != nil {
		rollbackErr = errors.Join(rollbackErr, u.authSvc.DisableUser(ctx, createdUser.AuthSubject))
	}

	return rollbackErr
}
