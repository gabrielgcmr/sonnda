package usersvc

import (
	"context"
	"errors"
	"strings"

	"sonnda-api/internal/app/apperr"

	userrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	professionalsvc "sonnda-api/internal/app/services/professional"
	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"
	external "sonnda-api/internal/domain/ports/integration"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type service struct {
	userRepo repository.User
	profRepo repository.Professional
	authSvc  external.IdentityService
}

var _ Service = (*service)(nil)

func New(
	userRepo repository.User,
	profRepo repository.Professional,
	authSvc external.IdentityService,
) Service {
	return &service{
		userRepo: userRepo,
		profRepo: profRepo,
		authSvc:  authSvc,
	}
}

func (s *service) Register(ctx context.Context, input UserRegisterInput) (*user.User, error) {
	if input.AccountType == user.AccountTypeProfessional {
		return s.createProfessionalUser(ctx, input)
	}
	return s.createUser(ctx, input)

}

func (s *service) createUser(ctx context.Context, input UserRegisterInput) (*user.User, error) {
	newUser, err := user.NewUser(user.NewUserParams{
		AuthProvider: input.Provider,
		AuthSubject:  input.Subject,
		Email:        input.Email,
		AccountType:  input.AccountType,
		FullName:     input.FullName,
		BirthDate:    input.BirthDate,
		CPF:          input.CPF,
		Phone:        input.Phone,
	})
	if err != nil {
		return nil, mapUserDomainError(err)
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		switch {
		case errors.Is(err, userrepo.ErrUserAlreadyExists):
			return nil, apperr.Conflict("usuário já cadastrado")
		default:
			return nil, err
		}
	}

	return newUser, nil
}

func (s *service) createProfessionalUser(ctx context.Context, input UserRegisterInput) (*user.User, error) {
	if input.Professional == nil {
		return nil, mapProfessionalDomainError(professional.ErrRegistrationRequired)
	}
	if s.profRepo == nil {
		return nil, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "serviço indisponível",
		}
	}

	kind := input.Professional.Kind.Normalize()
	if !kind.IsValid() {
		return nil, mapProfessionalDomainError(professional.ErrInvalidKind)
	}

	registrationNumber := strings.TrimSpace(input.Professional.RegistrationNumber)
	registrationIssuer := strings.TrimSpace(input.Professional.RegistrationIssuer)
	if registrationNumber == "" {
		return nil, mapProfessionalDomainError(professional.ErrInvalidRegistrationNumber)
	}
	if registrationIssuer == "" {
		return nil, mapProfessionalDomainError(professional.ErrInvalidRegistrationIssuer)
	}

	createdUser, err := s.createUser(ctx, input)
	if err != nil {
		return nil, err
	}

	profSvc := professionalsvc.New(s.profRepo)
	_, err = profSvc.Create(ctx, professionalsvc.CreateInput{
		UserID:             createdUser.ID,
		Kind:               kind,
		RegistrationNumber: registrationNumber,
		RegistrationIssuer: registrationIssuer,
		RegistrationState:  input.Professional.RegistrationState,
	})
	if err != nil {
		if rollbackErr := s.rollbackCreatedUser(ctx, createdUser); rollbackErr != nil {
			return nil, &apperr.AppError{
				Code:    apperr.INFRA_DATABASE_ERROR,
				Message: "falha tecnica",
				Cause:   errors.Join(err, rollbackErr),
			}
		}
		return nil, err
	}

	return createdUser, nil
}

func (s *service) rollbackCreatedUser(ctx context.Context, createdUser *user.User) error {
	if createdUser == nil {
		return nil
	}
	if s.userRepo != nil {
		if err := s.userRepo.SoftDelete(ctx, createdUser.ID); err != nil {
			return err
		}
	}

	if createdUser.AuthProvider == "firebase" && s.authSvc != nil {
		return s.authSvc.DisableUser(ctx, createdUser.AuthSubject)
	}

	return nil
}

func (s *service) Update(ctx context.Context, input UserUpdateInput) (*user.User, error) {
	existingUser, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, mapInfraError("userRepo.FindByID", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	changed, err := existingUser.ApplyUpdate(user.UpdateUserParams{
		FullName:  input.FullName,
		BirthDate: input.BirthDate,
		CPF:       input.CPF,
		Phone:     input.Phone,
	})
	if err != nil {
		return nil, mapUserDomainError(err)
	}
	if !changed {
		return existingUser, nil
	}

	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, mapInfraError("userRepo.Update", err)
	}

	return existingUser, nil
}

func mapInfraError(s string, err error) error {
	panic("unimplemented")
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID) error {
	existing, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *service) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	// 1) Carrega para validar existência (sem infra mapping)
	existing, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	// 2) Soft delete (idempotência: escolha A ou B)
	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		// A) Estrito: se já estava deletado (ou não existe), retorna not found
		// if errors.Is(err, userrepo.ErrNotFound) {
		// 	return ErrUserNotFound
		// }

		// B) Idempotente (recomendado): se já estava deletado, considere sucesso
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return nil
		}
		return err
	}

	return nil
}
