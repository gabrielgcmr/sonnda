package usersvc

import (
	"context"
	"errors"
	"strings"
	"time"

	"sonnda-api/internal/app/apperr"
	userport "sonnda-api/internal/app/ports/inbound/user"
	"sonnda-api/internal/app/ports/outbound/integrations"
	"sonnda-api/internal/app/ports/outbound/repositories"
	professionalsvc "sonnda-api/internal/app/services/professional"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"

	"github.com/google/uuid"
)

type service struct {
	userRepo repositories.UserRepository
	profRepo repositories.ProfessionalRepository
	authSvc  integrations.IdentityService
}

var _ userport.UserService = (*service)(nil)

func New(
	userRepo repositories.UserRepository,
	profRepo repositories.ProfessionalRepository,
	authSvc integrations.IdentityService,
) userport.UserService {
	return &service{
		userRepo: userRepo,
		profRepo: profRepo,
		authSvc:  authSvc,
	}
}

func (s *service) Register(ctx context.Context, input userport.RegisterInput) (*user.User, error) {
	if input.AccountType == user.AccountTypeProfessional {
		return s.createProfessionalUser(ctx, input)
	}
	return s.createUser(ctx, input)

}

func (s *service) createUser(ctx context.Context, input userport.RegisterInput) (*user.User, error) {
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

	existingByEmail, err := s.userRepo.FindByEmail(ctx, newUser.Email)
	if err != nil {
		return nil, mapInfraError("userRepo.FindByEmail", err)
	}
	if existingByEmail != nil {
		return nil, mapUserDomainError(user.ErrEmailAlreadyExists)
	}

	existingByCPF, err := s.userRepo.FindByCPF(ctx, newUser.CPF)
	if err != nil {
		return nil, mapInfraError("userRepo.FindByCPF", err)
	}
	if existingByCPF != nil {
		return nil, mapUserDomainError(user.ErrCPFAlreadyExists)
	}

	existingByAuth, err := s.userRepo.FindByAuthIdentity(ctx, newUser.AuthProvider, newUser.AuthSubject)
	if err != nil {
		return nil, mapInfraError("userRepo.FindByAuthIdentity", err)
	}
	if existingByAuth != nil {
		return nil, mapUserDomainError(user.ErrAuthIdentityAlreadyExists)
	}

	if err := s.userRepo.Save(ctx, newUser); err != nil {
		return nil, mapInfraError("userRepo.Save", err)
	}

	return newUser, nil
}

func (s *service) createProfessionalUser(ctx context.Context, input userport.RegisterInput) (*user.User, error) {
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

func (s *service) Update(ctx context.Context, input userport.UpdateInput) (*user.User, error) {
	existingUser, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, mapInfraError("userRepo.FindByID", err)
	}
	if existingUser == nil {
		return nil, mapUserDomainError(user.ErrUserNotFound)
	}

	if input.FullName != nil {
		name := strings.TrimSpace(*input.FullName)
		if name == "" {
			return nil, mapUserDomainError(user.ErrInvalidFullName)
		}
		existingUser.FullName = name
	}

	if input.BirthDate != nil {
		birthDate := input.BirthDate.UTC()
		if birthDate.IsZero() || birthDate.After(time.Now().UTC()) {
			return nil, mapUserDomainError(user.ErrInvalidBirthDate)
		}
		existingUser.BirthDate = birthDate
	}

	if input.CPF != nil {
		normalizedCPF := cleanDigits(*input.CPF)
		if normalizedCPF == "" || len(normalizedCPF) != 11 {
			return nil, mapUserDomainError(user.ErrInvalidCPF)
		}
		existingUser.CPF = normalizedCPF
	}

	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		if phone == "" {
			return nil, mapUserDomainError(user.ErrInvalidPhone)
		}
		existingUser.Phone = phone
	}

	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, mapInfraError("userRepo.Update", err)
	}

	return existingUser, nil
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID) error {
	existing, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return mapInfraError("userRepo.FindByID", err)
	}
	if existing == nil {
		return mapUserDomainError(user.ErrUserNotFound)
	}

	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		return err
	}

	if existing.AuthProvider == "firebase" && s.authSvc != nil {
		if err := s.authSvc.DisableUser(ctx, existing.AuthSubject); err != nil {
			return err
		}
	}

	return nil
}

func cleanDigits(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
