// internal/app/services/professional/service_test.go
package professionalsvc

import (
	"context"
	"errors"
	"testing"

	repoerr "sonnda-api/internal/adapters/outbound/persistence/postgres/repository"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/professional"

	"github.com/google/uuid"
)

type fakeRepo struct {
	createErr error
	findRes   *professional.Professional
	findErr   error
	deleteErr error
}

func (r *fakeRepo) Create(ctx context.Context, profile *professional.Professional) error {
	return r.createErr
}
func (r *fakeRepo) Update(ctx context.Context, profile *professional.Professional) error {
	panic("unused")
}
func (r *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error { return r.deleteErr }

func (r *fakeRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unused")
}

func (r *fakeRepo) FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error) {
	return r.findRes, r.findErr
}
func (r *fakeRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	return r.findRes, r.findErr
}
func (r *fakeRepo) FindByRegistration(ctx context.Context, registrationNumber, registrationIssuer string) (*professional.Professional, error) {
	panic("unused")
}
func (r *fakeRepo) FindByName(ctx context.Context, name string, limit, offset int) ([]*professional.Professional, error) {
	panic("unused")
}

func TestCreate_InvalidKind_ReturnsValidationFailed(t *testing.T) {
	svc := New(&fakeRepo{})

	_, err := svc.Create(context.Background(), CreateInput{
		UserID:             uuid.Must(uuid.NewV7()),
		Kind:               professional.Kind(""),
		RegistrationNumber: "CRM-123",
		RegistrationIssuer: "CRM",
	})

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.VALIDATION_FAILED {
		t.Fatalf("expected VALIDATION_FAILED, got %s", appErr.Code)
	}
}

func TestCreate_RepoError_ReturnsInfraDatabaseError(t *testing.T) {
	sentinel := errors.New("db down")
	svc := New(&fakeRepo{createErr: errors.Join(repoerr.ErrRepositoryFailure, sentinel)})

	_, err := svc.Create(context.Background(), CreateInput{
		UserID:             uuid.Must(uuid.NewV7()),
		Kind:               professional.KindDoctor,
		RegistrationNumber: "CRM-123",
		RegistrationIssuer: "CRM",
	})

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.INFRA_DATABASE_ERROR {
		t.Fatalf("expected INFRA_DATABASE_ERROR, got %s", appErr.Code)
	}
}

func TestGetByID_NotFound_ReturnsNotFound(t *testing.T) {
	svc := New(&fakeRepo{findRes: nil})

	_, err := svc.GetByID(context.Background(), uuid.Must(uuid.NewV7()))

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.NOT_FOUND {
		t.Fatalf("expected NOT_FOUND, got %s", appErr.Code)
	}
}

func TestDeleteByID_NotFound_ReturnsNotFound(t *testing.T) {
	svc := New(&fakeRepo{findRes: nil})

	err := svc.Delete(context.Background(), uuid.Must(uuid.NewV7()))

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.NOT_FOUND {
		t.Fatalf("expected NOT_FOUND, got %s", appErr.Code)
	}
}
