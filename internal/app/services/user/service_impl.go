package usersvc

import (
	"context"
	"errors"
	"strings"
	"time"

	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"
	"sonnda-api/internal/domain/ports/integrations"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

type service struct {
	userRepo repositories.UserRepository
	profRepo repositories.ProfessionalRepository
	authSvc  integrations.IdentityService
}

var _ Service = (*service)(nil)

func New(
	userRepo repositories.UserRepository,
	profRepo repositories.ProfessionalRepository,
	authSvc integrations.IdentityService,
) Service {
	return &service{
		userRepo: userRepo,
		profRepo: profRepo,
		authSvc:  authSvc,
	}
}

func (s *service) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	if input.Role == rbac.RoleDoctor || input.Role == rbac.RoleNurse {
		return s.createProfessionalUser(ctx, input)
	}
	return s.createUser(ctx, input)

}

func (s *service) createUser(ctx context.Context, input RegisterInput) (*user.User, error) {
	newUser, err := user.NewUser(user.NewUserParams{
		AuthProvider: input.Provider,
		AuthSubject:  input.Subject,
		Email:        input.Email,
		Role:         input.Role,
		FullName:     input.FullName,
		BirthDate:    input.BirthDate,
		CPF:          input.CPF,
		Phone:        input.Phone,
	})
	if err != nil {
		return nil, err
	}

	existingByEmail, err := s.userRepo.FindByEmail(ctx, newUser.Email)
	if err != nil {
		return nil, err
	}
	if existingByEmail != nil {
		return nil, user.ErrEmailAlreadyExists
	}

	existingByCPF, err := s.userRepo.FindByCPF(ctx, newUser.CPF)
	if err != nil {
		return nil, err
	}
	if existingByCPF != nil {
		return nil, user.ErrCPFAlreadyExists
	}

	existingByAuth, err := s.userRepo.FindByAuthIdentity(ctx, newUser.AuthProvider, newUser.AuthSubject)
	if err != nil {
		return nil, err
	}
	if existingByAuth != nil {
		return nil, user.ErrAuthIdentityAlreadyExists
	}

	if err := s.userRepo.Save(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *service) createProfessionalUser(ctx context.Context, input RegisterInput) (*user.User, error) {
	if input.Professional == nil {
		return nil, professional.ErrRegistrationRequired
	}
	if s.profRepo == nil {
		return nil, errors.New("professional repository not configured")
	}

	registrationNumber := strings.TrimSpace(input.Professional.RegistrationNumber)
	registrationIssuer := strings.TrimSpace(input.Professional.RegistrationIssuer)
	if registrationNumber == "" && registrationIssuer == "" {
		return nil, professional.ErrRegistrationRequired
	}
	if registrationNumber == "" {
		return nil, professional.ErrInvalidRegistrationNumber
	}
	if registrationIssuer == "" {
		return nil, professional.ErrInvalidRegistrationIssuer
	}

	createdUser, err := s.createUser(ctx, input)
	if err != nil {
		return nil, err
	}

	prof, err := professional.NewProfessional(professional.NewProfessionalParams{
		UserID:             createdUser.ID,
		RegistrationNumber: registrationNumber,
		RegistrationIssuer: registrationIssuer,
		RegistrationState:  input.Professional.RegistrationState,
	})
	if err != nil {
		return nil, err
	}

	if err := s.profRepo.Create(ctx, prof); err != nil {
		if rollbackErr := s.rollbackCreatedUser(ctx, createdUser); rollbackErr != nil {
			return nil, errors.Join(err, rollbackErr)
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

func (s *service) Update(ctx context.Context, input UpdateInput) (*user.User, error) {
	existingUser, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, user.ErrUserNotFound
	}

	if input.FullName != nil {
		name := strings.TrimSpace(*input.FullName)
		if name == "" {
			return nil, user.ErrInvalidFullName
		}
		existingUser.FullName = name
	}

	if input.BirthDate != nil {
		birthDate := input.BirthDate.UTC()
		if birthDate.IsZero() || birthDate.After(time.Now().UTC()) {
			return nil, user.ErrInvalidBirthDate
		}
		existingUser.BirthDate = birthDate
	}

	if input.CPF != nil {
		normalizedCPF := cleanDigits(*input.CPF)
		if normalizedCPF == "" || len(normalizedCPF) != 11 {
			return nil, user.ErrInvalidCPF
		}
		existingUser.CPF = normalizedCPF
	}

	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		if phone == "" {
			return nil, user.ErrInvalidPhone
		}
		existingUser.Phone = phone
	}

	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID) error {
	existing, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return user.ErrUserNotFound
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
