package patientsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/demographics"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type fakePatientRepo struct {
	findByCPFRes *patient.Patient
	findByCPFErr error

	findByIDRes *patient.Patient
	findByIDErr error

	createErr     error
	updateErr     error
	softDeleteErr error
	listErr       error

	calledFindByCPF bool
	calledFindByID  bool
	calledCreate    bool
	calledUpdate    bool
	calledSoftDel   bool
	calledList      bool
}

func (r *fakePatientRepo) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	r.calledFindByCPF = true
	return r.findByCPFRes, r.findByCPFErr
}

func (r *fakePatientRepo) Create(ctx context.Context, p *patient.Patient) error {
	r.calledCreate = true
	return r.createErr
}

func (r *fakePatientRepo) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	r.calledFindByID = true
	return r.findByIDRes, r.findByIDErr
}

func (r *fakePatientRepo) Update(ctx context.Context, p *patient.Patient) error {
	r.calledUpdate = true
	return r.updateErr
}

func (r *fakePatientRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	r.calledSoftDel = true
	return r.softDeleteErr
}

func (r *fakePatientRepo) List(ctx context.Context, limit, offset int) ([]patient.Patient, error) {
	r.calledList = true
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

func TestService_Create_CPFAlreadyExists(t *testing.T) {
	repo := &fakePatientRepo{findByCPFRes: &patient.Patient{ID: uuid.Must(uuid.NewV7())}}
	svc := New(repo, AllowAllPolicy{})

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
	if !errors.Is(err, patient.ErrCPFAlreadyExists) {
		t.Fatalf("expected ErrCPFAlreadyExists in chain, got %v", err)
	}
}

func TestService_Create_InvalidBirthDate(t *testing.T) {
	repo := &fakePatientRepo{}
	svc := New(repo, AllowAllPolicy{})

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
		t.Fatalf("expected shared.ErrInvalidBirthDate in chain, got %v", err)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := &fakePatientRepo{findByIDRes: nil}
	svc := New(repo, AllowAllPolicy{})

	_, err := svc.GetByID(context.Background(), &user.User{ID: uuid.Must(uuid.NewV7())}, uuid.Must(uuid.NewV7()))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var ae *apperr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apperr.AppError, got %T", err)
	}
	if ae.Code != apperr.NOT_FOUND {
		t.Fatalf("expected code %s, got %s", apperr.NOT_FOUND, ae.Code)
	}
	if !errors.Is(err, patient.ErrPatientNotFound) {
		t.Fatalf("expected ErrPatientNotFound in chain, got %v", err)
	}
}
