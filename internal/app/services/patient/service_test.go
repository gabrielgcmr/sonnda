package patientsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	repoerr "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/app/apperr"
	authorization "sonnda-api/internal/app/services/authorization"
	"sonnda-api/internal/domain/model/demographics"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type allowAllAuthorizer struct{}

func (a allowAllAuthorizer) Require(ctx context.Context, actor *user.User, action rbac.Action, patientID *uuid.UUID) error {
	return nil
}

var _ authorization.Authorizer = (*allowAllAuthorizer)(nil)

type fakePatientRepo struct {
	findByIDRes *patient.Patient
	findByIDErr error

	createErr     error
	updateErr     error
	softDeleteErr error
	listErr       error
}

func (r *fakePatientRepo) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	panic("unused")
}

func (r *fakePatientRepo) Create(ctx context.Context, p *patient.Patient) error { return r.createErr }
func (r *fakePatientRepo) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	return r.findByIDRes, r.findByIDErr
}
func (r *fakePatientRepo) Update(ctx context.Context, p *patient.Patient) error { return r.updateErr }
func (r *fakePatientRepo) SoftDelete(ctx context.Context, id uuid.UUID) error   { return r.softDeleteErr }
func (r *fakePatientRepo) List(ctx context.Context, limit, offset int) ([]patient.Patient, error) {
	return nil, r.listErr
}
func (r *fakePatientRepo) ListByName(ctx context.Context, name string, limit, offset int) ([]patient.Patient, error) {
	return nil, nil
}
func (r *fakePatientRepo) ListByBirthDate(ctx context.Context, birthDate time.Time, limit, offset int) ([]patient.Patient, error) {
	return nil, nil
}
func (r *fakePatientRepo) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]patient.Patient, error) {
	return nil, nil
}

func TestService_Create_AlreadyExists_ReturnsAlreadyExists(t *testing.T) {
	repo := &fakePatientRepo{createErr: repoerr.ErrPatientAlreadyExists}
	svc := New(repo, allowAllAuthorizer{})

	_, err := svc.Create(context.Background(), &user.User{ID: uuid.Must(uuid.NewV7())}, CreateInput{
		CPF:       "52998224725",
		FullName:  "Pessoa Teste",
		BirthDate: time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var ae *apperr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apperr.AppError, got %T", err)
	}
	if ae.Code != apperr.RESOURCE_ALREADY_EXISTS {
		t.Fatalf("expected code %s, got %s", apperr.RESOURCE_ALREADY_EXISTS, ae.Code)
	}
	if !errors.Is(err, repoerr.ErrPatientAlreadyExists) {
		t.Fatalf("expected ErrPatientAlreadyExists in chain, got %v", err)
	}
}

func TestService_Create_InvalidBirthDate_ReturnsValidationFailed(t *testing.T) {
	repo := &fakePatientRepo{}
	svc := New(repo, allowAllAuthorizer{})

	_, err := svc.Create(context.Background(), &user.User{ID: uuid.Must(uuid.NewV7())}, CreateInput{
		CPF:       "52998224725",
		FullName:  "Pessoa Teste",
		BirthDate: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var ae *apperr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apperr.AppError, got %T", err)
	}
	if ae.Code != apperr.VALIDATION_FAILED {
		t.Fatalf("expected code %s, got %s", apperr.VALIDATION_FAILED, ae.Code)
	}
	if !errors.Is(err, demographics.ErrInvalidBirthDate) {
		t.Fatalf("expected ErrInvalidBirthDate in chain, got %v", err)
	}
}

func TestService_GetByID_NotFound_ReturnsAccessDenied(t *testing.T) {
	repo := &fakePatientRepo{findByIDRes: nil}
	svc := New(repo, allowAllAuthorizer{})

	_, err := svc.GetByID(context.Background(), &user.User{ID: uuid.Must(uuid.NewV7())}, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var ae *apperr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apperr.AppError, got %T", err)
	}
	if ae.Code != apperr.ACCESS_DENIED {
		t.Fatalf("expected code %s, got %s", apperr.ACCESS_DENIED, ae.Code)
	}
}

func TestService_Create_RepoFailure_ReturnsInfraDatabaseError(t *testing.T) {
	sentinel := errors.New("db down")
	repo := &fakePatientRepo{createErr: errors.Join(repoerr.ErrRepositoryFailure, sentinel)}
	svc := New(repo, allowAllAuthorizer{})

	_, err := svc.Create(context.Background(), &user.User{ID: uuid.Must(uuid.NewV7())}, CreateInput{
		CPF:       "52998224725",
		FullName:  "Pessoa Teste",
		BirthDate: time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC),
	})

	var ae *apperr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apperr.AppError, got %T", err)
	}
	if ae.Code != apperr.INFRA_DATABASE_ERROR {
		t.Fatalf("expected code %s, got %s", apperr.INFRA_DATABASE_ERROR, ae.Code)
	}
}
