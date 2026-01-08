package registrationsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"sonnda-api/internal/app/apperr"
	professionalsvc "sonnda-api/internal/app/services/professional"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type fakeUserSvc struct {
	createRes *user.User
	createErr error

	softDeleteErr error
	softDeleteN   int
}

func (s *fakeUserSvc) Create(ctx context.Context, input usersvc.UserCreateInput) (*user.User, error) {
	return s.createRes, s.createErr
}
func (s *fakeUserSvc) Update(ctx context.Context, input usersvc.UserUpdateInput) (*user.User, error) {
	panic("unused")
}
func (s *fakeUserSvc) Delete(ctx context.Context, userID uuid.UUID) error { panic("unused") }
func (s *fakeUserSvc) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	s.softDeleteN++
	return s.softDeleteErr
}

type fakeProfSvc struct {
	createErr error
}

func (s *fakeProfSvc) Create(ctx context.Context, input professionalsvc.CreateInput) (*professional.Professional, error) {
	return nil, s.createErr
}
func (s *fakeProfSvc) GetByID(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error) {
	panic("unused")
}
func (s *fakeProfSvc) GetByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	panic("unused")
}
func (s *fakeProfSvc) Delete(ctx context.Context, profileID uuid.UUID) error {
	panic("unused")
}

type fakeIdentitySvc struct {
	disableErr error
}

func (s *fakeIdentitySvc) ProviderName() string { return "fake" }
func (s *fakeIdentitySvc) VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error) {
	panic("unused")
}
func (s *fakeIdentitySvc) DisableUser(ctx context.Context, subject string) error { return s.disableErr }

func TestRegister_BasicUser_SkipsProfessional(t *testing.T) {
	created := &user.User{ID: uuid.Must(uuid.NewV7())}
	userSvc := &fakeUserSvc{createRes: created}

	svc := New(userSvc, &fakeProfSvc{createErr: errors.New("should not be called")}, nil)

	out, err := svc.Register(context.Background(), RegisterInput{
		Provider:    "firebase",
		Subject:     "sub",
		Email:       "a@b.com",
		AccountType: user.AccountTypeBasicCare,
		FullName:    "Nome",
		BirthDate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		CPF:         "12345678901",
		Phone:       "11999999999",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out != created {
		t.Fatalf("expected created user")
	}
	if userSvc.softDeleteN != 0 {
		t.Fatalf("expected no rollback")
	}
}

func TestRegister_Professional_MissingProfessionalInput_ReturnsValidation(t *testing.T) {
	created := &user.User{ID: uuid.Must(uuid.NewV7())}
	userSvc := &fakeUserSvc{createRes: created}

	svc := New(userSvc, &fakeProfSvc{}, nil)

	_, err := svc.Register(context.Background(), RegisterInput{
		Provider:     "firebase",
		Subject:      "sub",
		Email:        "a@b.com",
		AccountType:  user.AccountTypeProfessional,
		FullName:     "Nome",
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		CPF:          "12345678901",
		Phone:        "11999999999",
		Professional: nil,
	})

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.VALIDATION_FAILED {
		t.Fatalf("expected VALIDATION_FAILED, got %s", appErr.Code)
	}
}

func TestRegister_Professional_ProfCreateFails_RollsBack(t *testing.T) {
	created := &user.User{
		ID:           uuid.Must(uuid.NewV7()),
		AuthProvider: "firebase",
		AuthSubject:  "sub",
	}
	userSvc := &fakeUserSvc{createRes: created}
	profErr := apperr.Validation("dados profissionais inválidos")

	svc := New(userSvc, &fakeProfSvc{createErr: profErr}, &fakeIdentitySvc{})

	_, err := svc.Register(context.Background(), RegisterInput{
		Provider:    "firebase",
		Subject:     "sub",
		Email:       "a@b.com",
		AccountType: user.AccountTypeProfessional,
		FullName:    "Nome",
		BirthDate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		CPF:         "12345678901",
		Phone:       "11999999999",
		Professional: &ProfessionalInput{
			Kind:               professional.KindDoctor,
			RegistrationNumber: "CRM-123",
			RegistrationIssuer: "CRM",
		},
	})

	if !errors.Is(err, profErr) {
		t.Fatalf("expected original professional error, got %v", err)
	}
	if userSvc.softDeleteN != 1 {
		t.Fatalf("expected rollback soft delete once, got %d", userSvc.softDeleteN)
	}
}
