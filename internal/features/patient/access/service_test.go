// internal/features/patient/access/service_test.go
package patientaccess

import (
	"context"
	"errors"
	"testing"

	domainrepository "github.com/gabrielgcmr/sonnda/internal/domain/repository"
	accessdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/access/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type listRepository struct {
	patients  []AccessiblePatient
	total     int64
	err       error
	accountID uuid.UUID
	limit     int
	offset    int
}

func (r *listRepository) ListAccessiblePatientsByUser(_ context.Context, accountID uuid.UUID, limit, offset int) ([]AccessiblePatient, int64, error) {
	r.accountID, r.limit, r.offset = accountID, limit, offset
	return r.patients, r.total, r.err
}

func (*listRepository) Upsert(context.Context, *accessdomain.PatientAccess) error { return nil }
func (*listRepository) HasActiveAccess(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

func TestListForAccountOmitsRelationshipMetadata(t *testing.T) {
	accountID, patientID := uuid.New(), uuid.New()
	avatarURL := "https://example.test/avatar.png"
	repo := &listRepository{
		patients: []AccessiblePatient{{
			PatientID:    patientID,
			FullName:     "Ana Silva",
			AvatarURL:    &avatarURL,
			RelationType: "professional",
		}},
		total: 1,
	}

	output, err := NewService(repo).ListForAccount(context.Background(), accountID, 150, 2)
	if err != nil {
		t.Fatal(err)
	}
	if repo.accountID != accountID || repo.limit != 100 || repo.offset != 2 {
		t.Fatalf("unexpected repository input: account=%s limit=%d offset=%d", repo.accountID, repo.limit, repo.offset)
	}
	if output.Total != 1 || output.Limit != 100 || output.Offset != 2 || len(output.Patients) != 1 {
		t.Fatalf("unexpected output: %+v", output)
	}
	patient := output.Patients[0]
	if patient.ID != patientID || patient.FullName != "Ana Silva" || patient.AvatarURL == nil || *patient.AvatarURL != avatarURL {
		t.Fatalf("unexpected patient summary: %+v", patient)
	}
}

func TestListForAccountMapsRepositoryFailure(t *testing.T) {
	dbErr := errors.New("database unavailable")
	err := errors.Join(domainrepository.ErrRepositoryFailure, dbErr)
	_, got := NewService(&listRepository{err: err}).ListForAccount(context.Background(), uuid.New(), 20, 0)
	var appErr *apperr.AppError
	if !errors.As(got, &appErr) || appErr.Kind != apperr.INFRA_DATABASE_ERROR || !errors.Is(got, dbErr) {
		t.Fatalf("unexpected error: %v", got)
	}
}
