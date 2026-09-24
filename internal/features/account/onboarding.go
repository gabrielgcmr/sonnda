// internal/features/account/onboarding.go
package account

import (
	"context"
	"errors"

	professionalsvc "github.com/gabrielgcmr/sonnda/internal/application/services/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type Onboarding interface {
	Register(ctx context.Context, input RegisterInput) (*user.User, error)
}

type onboarding struct {
	userRepo repository.User
	userSvc  Service
	profSvc  professionalsvc.Service
}

var _ Onboarding = (*onboarding)(nil)

func NewOnboarding(userRepo repository.User, userSvc Service, profSvc professionalsvc.Service) Onboarding {
	return &onboarding{
		userRepo: userRepo,
		userSvc:  userSvc,
		profSvc:  profSvc,
	}
}

func (u *onboarding) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	if input.AccountType == user.AccountTypeProfessional {
		return nil, apperr.DomainRuleViolation("criação de profissional ainda não está implementada no MVP")
	}

	// Verificar se usuário já existe
	existing, err := u.userRepo.FindByAuthIdentity(ctx, input.Issuer, input.Subject)
	if err != nil {
		return nil, apperr.Internal("falha ao verificar registro", err)
	}
	if existing != nil {
		return nil, apperr.AlreadyExists("usuário já cadastrado")
	}

	createdUser, err := u.userSvc.Create(ctx, UserCreateInput{
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
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr != nil {
			return nil, appErr
		}
		return nil, apperr.Internal("falha ao criar usuário", err)
	}

	return createdUser, nil
}
