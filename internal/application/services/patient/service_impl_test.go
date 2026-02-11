// internal/application/services/patient/service_impl_test.go
package patientsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patientaccess"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/rbac"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type allowAllAuthorizer struct{}

func (a allowAllAuthorizer) Require(ctx context.Context, actor *user.User, action rbac.Action, patientID *uuid.UUID) error {
	return nil
}

type fakePatientRepo struct {
	created   *patient.Patient
	createErr error
}

func (r *fakePatientRepo) Create(ctx context.Context, p *patient.Patient) error {
	r.created = p
	return r.createErr
}
func (r *fakePatientRepo) Update(ctx context.Context, p *patient.Patient) error { panic("unused") }
func (r *fakePatientRepo) SoftDelete(ctx context.Context, id uuid.UUID) error   { panic("unused") }
func (r *fakePatientRepo) HardDelete(ctx context.Context, id uuid.UUID) error   { panic("unused") }
func (r *fakePatientRepo) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	panic("unused")
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

type fakeAccessRepo struct {
	upsertAccess *patientaccess.PatientAccess
	upsertErr    error
}

func (r *fakeAccessRepo) ListAccessiblePatientsByUser(
	ctx context.Context,
	granteeID uuid.UUID,
	limit, offset int,
) ([]repository.AccessiblePatient, int64, error) {
	panic("unused")
}

func (r *fakeAccessRepo) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	r.upsertAccess = access
	return r.upsertErr
}

func (r *fakeAccessRepo) HasActiveAccess(ctx context.Context, patientID, granteeID uuid.UUID) (bool, error) {
	panic("unused")
}

func TestCreate_ProfessionalCreatesAccess(t *testing.T) {
	patientRepo := &fakePatientRepo{}
	accessRepo := &fakeAccessRepo{}
	svc := New(patientRepo, accessRepo, allowAllAuthorizer{})

	currentUser := &user.User{
		ID:          uuid.Must(uuid.NewV7()),
		AccountType: user.AccountTypeProfessional,
	}

	input := CreateInput{
		CPF:       "12345678901",
		FullName:  "Joana Silva",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:    demographics.GenderFemale,
		Race:      demographics.RaceWhite,
		AvatarURL: "https://example.com/avatar.png",
	}

	created, err := svc.Create(context.Background(), currentUser, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || patientRepo.created == nil {
		t.Fatalf("expected patient to be created")
	}
	if accessRepo.upsertAccess == nil {
		t.Fatalf("expected patient access to be created")
	}
	if accessRepo.upsertAccess.PatientID != created.ID {
		t.Fatalf("expected patient_id=%s, got %s", created.ID, accessRepo.upsertAccess.PatientID)
	}
	if accessRepo.upsertAccess.GranteeID != currentUser.ID {
		t.Fatalf("expected grantee_id=%s, got %s", currentUser.ID, accessRepo.upsertAccess.GranteeID)
	}
	if accessRepo.upsertAccess.GrantedBy == nil || *accessRepo.upsertAccess.GrantedBy != currentUser.ID {
		t.Fatalf("expected granted_by=%s", currentUser.ID)
	}
	if accessRepo.upsertAccess.RelationType != patientaccess.RelationshipTypeProfessional {
		t.Fatalf("expected relation_type=%s, got %s", patientaccess.RelationshipTypeProfessional, accessRepo.upsertAccess.RelationType)
	}
}

func TestCreate_BasicCareCreatesAccess(t *testing.T) {
	accessRepo := &fakeAccessRepo{}
	svc := New(&fakePatientRepo{}, accessRepo, allowAllAuthorizer{})

	currentUser := &user.User{
		ID:          uuid.Must(uuid.NewV7()),
		AccountType: user.AccountTypeBasicCare,
	}

	input := CreateInput{
		CPF:       "12345678901",
		FullName:  "Joana Silva",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:    demographics.GenderFemale,
		Race:      demographics.RaceWhite,
		AvatarURL: "https://example.com/avatar.png",
	}

	_, err := svc.Create(context.Background(), currentUser, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessRepo.upsertAccess == nil {
		t.Fatalf("expected patient access to be created")
	}
	if accessRepo.upsertAccess.RelationType != patientaccess.RelationshipTypeCaregiver {
		t.Fatalf("expected relation_type=%s, got %s", patientaccess.RelationshipTypeCaregiver, accessRepo.upsertAccess.RelationType)
	}
}

func TestCreate_AccessRepoError_ReturnsInfraDatabaseError(t *testing.T) {
	svc := New(&fakePatientRepo{}, &fakeAccessRepo{upsertErr: errors.New("db down")}, allowAllAuthorizer{})

	currentUser := &user.User{
		ID:          uuid.Must(uuid.NewV7()),
		AccountType: user.AccountTypeProfessional,
	}

	input := CreateInput{
		CPF:       "12345678901",
		FullName:  "Joana Silva",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:    demographics.GenderFemale,
		Race:      demographics.RaceWhite,
		AvatarURL: "https://example.com/avatar.png",
	}

	_, err := svc.Create(context.Background(), currentUser, input)

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Kind != apperr.INFRA_DATABASE_ERROR {
		t.Fatalf("expected INFRA_DATABASE_ERROR, got %s", appErr.Kind)
	}
}

func TestCreate_NilUser_ReturnsAuthRequired(t *testing.T) {
	svc := New(&fakePatientRepo{}, &fakeAccessRepo{}, allowAllAuthorizer{})

	input := CreateInput{
		CPF:       "12345678901",
		FullName:  "Joana Silva",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:    demographics.GenderFemale,
		Race:      demographics.RaceWhite,
		AvatarURL: "https://example.com/avatar.png",
	}

	_, err := svc.Create(context.Background(), nil, input)

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Kind != apperr.AUTH_REQUIRED {
		t.Fatalf("expected AUTH_REQUIRED, got %s", appErr.Kind)
	}
}
