// internal/application/services/labs/service_impl_test.go
package labsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type fakePatientRepo struct {
	findByIDRes *patient.Patient
	findByIDErr error
}

func (r *fakePatientRepo) Create(ctx context.Context, p *patient.Patient) error { panic("unused") }
func (r *fakePatientRepo) Update(ctx context.Context, p *patient.Patient) error { panic("unused") }
func (r *fakePatientRepo) SoftDelete(ctx context.Context, id uuid.UUID) error   { panic("unused") }
func (r *fakePatientRepo) HardDelete(ctx context.Context, id uuid.UUID) error   { panic("unused") }
func (r *fakePatientRepo) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	return r.findByIDRes, r.findByIDErr
}
func (r *fakePatientRepo) FindByName(ctx context.Context, name string) ([]patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) List(ctx context.Context, limit, offset int) ([]patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) SearchByName(ctx context.Context, name string, limit, offset int) ([]patient.Patient, error) {
	panic("unused")
}

type fakeLabsRepo struct {
	listRes []labs.LabReport
	listErr error
}

func (r *fakeLabsRepo) Create(ctx context.Context, report *labs.LabReport) error { panic("unused") }
func (r *fakeLabsRepo) ExistsBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (bool, error) {
	panic("unused")
}
func (r *fakeLabsRepo) Delete(ctx context.Context, id uuid.UUID) error { panic("unused") }
func (r *fakeLabsRepo) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	panic("unused")
}
func (r *fakeLabsRepo) ListLabs(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]labs.LabReport, error) {
	return r.listRes, r.listErr
}
func (r *fakeLabsRepo) ListItemsByPatientAndParameter(
	ctx context.Context,
	patientID uuid.UUID,
	parameterName string,
	limit, offset int,
) ([]labs.LabResultItemTimeline, error) {
	panic("unused")
}

func TestList_InvalidPatientID_ReturnsValidationFailed(t *testing.T) {
	svc := New(&fakePatientRepo{}, &fakeLabsRepo{})

	_, err := svc.List(context.Background(), uuid.Nil, 10, 0)

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Kind != apperr.VALIDATION_FAILED {
		t.Fatalf("expected VALIDATION_FAILED, got %s", appErr.Kind)
	}
}

func TestList_PatientNotFound_ReturnsNotFound(t *testing.T) {
	svc := New(&fakePatientRepo{findByIDRes: nil}, &fakeLabsRepo{})

	_, err := svc.List(context.Background(), uuid.Must(uuid.NewV7()), 10, 0)

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Kind != apperr.NOT_FOUND {
		t.Fatalf("expected NOT_FOUND, got %s", appErr.Kind)
	}
}

func TestList_PatientRepoError_ReturnsInfraDatabaseError(t *testing.T) {
	sentinel := errors.New("db down")
	svc := New(&fakePatientRepo{findByIDErr: errors.Join(repo.ErrRepositoryFailure, sentinel)}, &fakeLabsRepo{})

	_, err := svc.List(context.Background(), uuid.Must(uuid.NewV7()), 10, 0)

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Kind != apperr.INFRA_DATABASE_ERROR {
		t.Fatalf("expected INFRA_DATABASE_ERROR, got %s", appErr.Kind)
	}
}

func TestList_LabsRepoError_ReturnsInfraDatabaseError(t *testing.T) {
	svc := New(
		&fakePatientRepo{findByIDRes: &patient.Patient{ID: uuid.Must(uuid.NewV7())}},
		&fakeLabsRepo{listErr: errors.New("db down")},
	)

	_, err := svc.List(context.Background(), uuid.Must(uuid.NewV7()), 10, 0)

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Kind != apperr.INFRA_DATABASE_ERROR {
		t.Fatalf("expected INFRA_DATABASE_ERROR, got %s", appErr.Kind)
	}
}
