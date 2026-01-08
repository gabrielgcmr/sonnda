package registrationsvc

import (
	"context"
	"errors"

	"sonnda-api/internal/app/apperr"
	professionalsvc "sonnda-api/internal/app/services/professional"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/model/user"
	external "sonnda-api/internal/domain/ports/integration"
)

type service struct {
	userSvc usersvc.Service
	profSvc professionalsvc.Service
	authSvc external.IdentityService
}

var _ Service = (*service)(nil)

func New(userSvc usersvc.Service, profSvc professionalsvc.Service, authSvc external.IdentityService) Service {
	return &service{
		userSvc: userSvc,
		profSvc: profSvc,
		authSvc: authSvc,
	}
}

func (s *service) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	if s == nil || s.userSvc == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("user service not configured"))
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

	if s.profSvc == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("professional service not configured"))
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
	if s.userSvc != nil {
		rollbackErr = errors.Join(rollbackErr, s.userSvc.SoftDelete(ctx, createdUser.ID))
	}

	if createdUser.AuthProvider == "firebase" && s.authSvc != nil {
		rollbackErr = errors.Join(rollbackErr, s.authSvc.DisableUser(ctx, createdUser.AuthSubject))
	}

	return rollbackErr
}
