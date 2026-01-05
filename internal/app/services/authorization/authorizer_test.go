package authorization

import (
	"context"
	"errors"
	"testing"
	"time"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/app/interfaces/repositories"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/patient/patientaccess"
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"

	"github.com/google/uuid"
)

type fakePatientRepo struct {
	findByIDRes *patient.Patient
	findByIDErr error
}

func (r *fakePatientRepo) Create(ctx context.Context, patient *patient.Patient) error {
	panic("unused")
}
func (r *fakePatientRepo) Update(ctx context.Context, patient *patient.Patient) error {
	panic("unused")
}
func (r *fakePatientRepo) SoftDelete(ctx context.Context, id uuid.UUID) error { panic("unused") }
func (r *fakePatientRepo) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	return r.findByIDRes, r.findByIDErr
}
func (r *fakePatientRepo) List(ctx context.Context, limit, offset int) ([]patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) ListByName(ctx context.Context, name string, limit, offset int) ([]patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) ListByBirthDate(ctx context.Context, birthDate time.Time, limit, offset int) ([]patient.Patient, error) {
	panic("unused")
}
func (r *fakePatientRepo) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]patient.Patient, error) {
	panic("unused")
}

type fakePatientAccessRepo struct {
	findRes *patientaccess.PatientAccess
	findErr error
}

func (r *fakePatientAccessRepo) Find(ctx context.Context, patientID, userID uuid.UUID) (*patientaccess.PatientAccess, error) {
	return r.findRes, r.findErr
}
func (r *fakePatientAccessRepo) FindActive(ctx context.Context, patientID, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unused")
}
func (r *fakePatientAccessRepo) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unused")
}
func (r *fakePatientAccessRepo) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	panic("unused")
}

type fakeProfRepo struct {
	findRes *professional.Professional
	findErr error
}

func (r *fakeProfRepo) Create(ctx context.Context, profile *professional.Professional) error {
	panic("unused")
}
func (r *fakeProfRepo) Update(ctx context.Context, profile *professional.Professional) error {
	panic("unused")
}
func (r *fakeProfRepo) Delete(ctx context.Context, id uuid.UUID) error { panic("unused") }
func (r *fakeProfRepo) FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error) {
	panic("unused")
}
func (r *fakeProfRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	return r.findRes, r.findErr
}
func (r *fakeProfRepo) FindByRegistration(ctx context.Context, registrationNumber, registrationIssuer string) (*professional.Professional, error) {
	panic("unused")
}
func (r *fakeProfRepo) FindByName(ctx context.Context, name string, limit, offset int) ([]*professional.Professional, error) {
	panic("unused")
}

func TestRequire_RBACDenied_ReturnsActionNotAllowed(t *testing.T) {
	svc := New(&fakePatientRepo{}, &fakePatientAccessRepo{}, &fakeProfRepo{})

	actor := &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare}
	err := svc.Require(context.Background(), actor, rbac.ActionWriteClinicalNote, nil)
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.ACTION_NOT_ALLOWED {
		t.Fatalf("expected ACTION_NOT_ALLOWED, got %s", appErr.Code)
	}
}

func TestRequire_PatientNotFound_ReturnsAccessDenied(t *testing.T) {
	patientID := uuid.Must(uuid.NewV7())
	svc := New(&fakePatientRepo{findByIDRes: nil}, &fakePatientAccessRepo{}, &fakeProfRepo{})

	actor := &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare}
	err := svc.Require(context.Background(), actor, rbac.ActionReadPatient, &patientID)
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperr.ACCESS_DENIED {
		t.Fatalf("expected ACCESS_DENIED, got %s", appErr.Code)
	}
}

func TestRequire_PatientOwner_Allows(t *testing.T) {
	actorID := uuid.Must(uuid.NewV7())
	patientID := uuid.Must(uuid.NewV7())
	ownerID := actorID
	svc := New(
		&fakePatientRepo{findByIDRes: &patient.Patient{ID: patientID, OwnerUserID: &ownerID}},
		&fakePatientAccessRepo{},
		&fakeProfRepo{},
	)

	actor := &user.User{ID: actorID, AccountType: user.AccountTypeBasicCare}
	if err := svc.Require(context.Background(), actor, rbac.ActionReadPatient, &patientID); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRequire_PatientAccessLink_Allows(t *testing.T) {
	actorID := uuid.Must(uuid.NewV7())
	patientID := uuid.Must(uuid.NewV7())
	svc := New(
		&fakePatientRepo{findByIDRes: &patient.Patient{ID: patientID}},
		&fakePatientAccessRepo{findRes: &patientaccess.PatientAccess{}},
		&fakeProfRepo{},
	)

	actor := &user.User{ID: actorID, AccountType: user.AccountTypeBasicCare}
	if err := svc.Require(context.Background(), actor, rbac.ActionReadPatient, &patientID); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRequire_WritePrescriptions_LoadsProfessionalKind(t *testing.T) {
	actorID := uuid.Must(uuid.NewV7())
	patientID := uuid.Must(uuid.NewV7())
	kind := professional.KindDoctor
	svc := New(
		&fakePatientRepo{findByIDRes: &patient.Patient{ID: patientID}},
		&fakePatientAccessRepo{findRes: &patientaccess.PatientAccess{}},
		&fakeProfRepo{findRes: &professional.Professional{UserID: actorID, Kind: kind}},
	)

	actor := &user.User{ID: actorID, AccountType: user.AccountTypeProfessional}
	if err := svc.Require(context.Background(), actor, rbac.ActionWritePrescriptions, &patientID); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

var _ repositories.PatientRepository = (*fakePatientRepo)(nil)
var _ repositories.PatientAccessRepository = (*fakePatientAccessRepo)(nil)
var _ repositories.ProfessionalRepository = (*fakeProfRepo)(nil)
