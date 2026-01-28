package registration

import (
	"context"

	professionalsvc "github.com/gabrielgcmr/sonnda/internal/app/services/professional"
	usersvc "github.com/gabrielgcmr/sonnda/internal/app/services/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/data"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

type UseCase interface {
	Register(ctx context.Context, input RegisterInput) (*user.User, error)
}

type usecase struct {
	userRepo     data.UserRepo
	userSvc      usersvc.Service
	profSvc      professionalsvc.Service
	authProvider security.IdentityProvider
}

var _ UseCase = (*usecase)(nil)

func New(userRepo data.UserRepo, userSvc usersvc.Service, profSvc professionalsvc.Service, authProvider security.IdentityProvider) *usecase {
	return &usecase{
		userRepo:     userRepo,
		userSvc:      userSvc,
		profSvc:      profSvc,
		authProvider: authProvider,
	}
}

func (u *usecase) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	// Verificar se usuário já existe
	existing, err := u.userRepo.FindByAuthIdentity(ctx, input.Issuer, input.Subject)
	if err != nil {
		return nil, apperr.Internal("falha ao verificar registro", err)
	}
	if existing != nil {
		return nil, &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "usuário já cadastrado",
		}
	}

	createdUser, err := u.userSvc.Create(ctx, usersvc.UserCreateInput{
		Issuer:      input.Issuer,
		Subject:     input.Subject,
		Email:       input.Email,
		AccountType: input.AccountType,
		FullName:    input.FullName,
		BirthDate:   input.BirthDate,
		CPF:         input.CPF,
		Phone:       input.Phone,
	})
	if err != nil {
		return nil, apperr.Internal("falha ao criar usuário", err)
	}

	if input.AccountType != user.AccountTypeProfessional {
		return createdUser, nil
	}

	if input.Professional == nil {
		return nil, apperr.Validation("dados do profissional inválidos")
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

	return nil, err
}
