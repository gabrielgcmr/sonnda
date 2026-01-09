package registrationsvc

import (
	"context"
	"errors"

	"sonnda-api/internal/app/apperr"
	professionalsvc "sonnda-api/internal/app/services/professional"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/ports/integration"
	"sonnda-api/internal/domain/ports/repository"
)

type service struct {
	userRepo repository.User
	userSvc  usersvc.Service
	profSvc  professionalsvc.Service
	authSvc  integration.IdentityService
}

var _ Service = (*service)(nil)

func New(userRepo repository.User, userSvc usersvc.Service, profSvc professionalsvc.Service, authSvc integration.IdentityService) Service {
	return &service{
		userRepo: userRepo,
		userSvc:  userSvc,
		profSvc:  profSvc,
		authSvc:  authSvc,
	}
}

func (s *service) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	// Verificar se usuário já existe
	existing, err := s.userRepo.FindByAuthIdentity(ctx, input.Provider, input.Subject)
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
			Message: "usuário já cadastrado",
		}
	}

	createdUser, err := s.userSvc.Create(ctx, usersvc.UserCreateInput{
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
		return nil, apperr.Validation("dados do profissional inválidos")
	}

	_, err = s.profSvc.Create(ctx, professionalsvc.CreateInput{
		UserID:             createdUser.ID,
		Kind:               input.Professional.Kind,
		RegistrationNumber: input.Professional.RegistrationNumber,
		RegistrationIssuer: input.Professional.RegistrationIssuer,
		RegistrationState:  input.Professional.RegistrationState,
	})
	if err == nil {
		return createdUser, nil
	}

	rollbackErr := s.rollbackCreatedUser(ctx, createdUser)
	if rollbackErr != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   errors.Join(err, rollbackErr),
		}
	}

	return nil, err
}

func (s *service) rollbackCreatedUser(ctx context.Context, createdUser *user.User) error {
	if createdUser == nil {
		return nil
	}

	var rollbackErr error

	if createdUser.AuthProvider == "firebase" && s.authSvc != nil {
		rollbackErr = errors.Join(rollbackErr, s.authSvc.DisableUser(ctx, createdUser.AuthSubject))
	}

	return rollbackErr
}
