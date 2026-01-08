package usersvc

import (
	"context"
	"errors"
	"testing"
	"time"

	userrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type fakeUserRepo struct {
	createErr error
	updateErr error
	deleteErr error

	findByIDRes *user.User
	findByIDErr error
}

func (r *fakeUserRepo) Create(ctx context.Context, u *user.User) error { return r.createErr }
func (r *fakeUserRepo) Update(ctx context.Context, u *user.User) error { return r.updateErr }
func (r *fakeUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return r.deleteErr }
func (r *fakeUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unused")
}
func (r *fakeUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return r.findByIDRes, r.findByIDErr
}
func (r *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	panic("unused")
}
func (r *fakeUserRepo) FindByCPF(ctx context.Context, cpf string) (*user.User, error) {
	panic("unused")
}
func (r *fakeUserRepo) FindByAuthIdentity(ctx context.Context, provider, subject string) (*user.User, error) {
	panic("unused")
}

func TestCreate_RepoAlreadyExists_ReturnsAlreadyExists(t *testing.T) {
	repo := &fakeUserRepo{createErr: userrepo.ErrUserAlreadyExists}
	svc := New(repo)

	_, err := svc.Create(context.Background(), UserCreateInput{
		Provider:    "firebase",
		Subject:     "sub",
		Email:       "a@b.com",
		FullName:    "Nome",
		AccountType: user.AccountTypeBasicCare,
		BirthDate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		CPF:         "12345678901",
		Phone:       "11999999999",
	})

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.RESOURCE_ALREADY_EXISTS {
		t.Fatalf("expected RESOURCE_ALREADY_EXISTS, got %s", appErr.Code)
	}
}

func TestUpdate_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeUserRepo{findByIDRes: nil, findByIDErr: nil}
	svc := New(repo)

	_, err := svc.Update(context.Background(), UserUpdateInput{UserID: uuid.Must(uuid.NewV7())})

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.NOT_FOUND {
		t.Fatalf("expected NOT_FOUND, got %s", appErr.Code)
	}
}

func TestUpdate_RepoFailure_ReturnsInfraDatabaseError(t *testing.T) {
	existing, err := user.NewUser(user.NewUserParams{
		AuthProvider: "firebase",
		AuthSubject:  "sub",
		Email:        "a@b.com",
		FullName:     "Nome",
		AccountType:  user.AccountTypeBasicCare,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		CPF:          "12345678901",
		Phone:        "11999999999",
	})
	if err != nil {
		t.Fatalf("setup user: %v", err)
	}

	repo := &fakeUserRepo{
		findByIDRes: existing,
		updateErr:   errors.Join(userrepo.ErrRepositoryFailure, errors.New("db down")),
	}
	svc := New(repo)

	newName := "Outro Nome"
	_, err = svc.Update(context.Background(), UserUpdateInput{
		UserID:   existing.ID,
		FullName: &newName,
	})

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.INFRA_DATABASE_ERROR {
		t.Fatalf("expected INFRA_DATABASE_ERROR, got %s", appErr.Code)
	}
}
