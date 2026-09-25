// internal/application/services/authorization/authorizer_test.go
package authorization

import (
	"context"
	"errors"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type patientLookup struct {
	patientprofile.Repository
	result *profiledomain.Patient
	err    error
	id     uuid.UUID
}

func (r *patientLookup) FindByID(_ context.Context, id uuid.UUID) (*profiledomain.Patient, error) {
	r.id = id
	return r.result, r.err
}

type accessLookup struct {
	repository.PatientAccessRepo
	allowed   bool
	err       error
	patientID uuid.UUID
	actorID   uuid.UUID
}

func (r *accessLookup) HasActiveAccess(_ context.Context, patientID, actorID uuid.UUID) (bool, error) {
	r.patientID, r.actorID = patientID, actorID
	return r.allowed, r.err
}

func TestRequirePatientAccess(t *testing.T) {
	actorID, otherID, patientID := uuid.New(), uuid.New(), uuid.New()
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name             string
		owner            *uuid.UUID
		accountType      accountdomain.AccountType
		allowed          bool
		missing          bool
		patientErr       error
		accessErr        error
		wantKind         apperr.ErrorKind
		wantAccessLookup bool
	}{
		{name: "owner", owner: &actorID},
		{name: "professional with grant", owner: &otherID, accountType: accountdomain.AccountTypeProfessional, allowed: true, wantAccessLookup: true},
		{name: "basic care with grant", accountType: accountdomain.AccountTypeBasicCare, allowed: true, wantAccessLookup: true},
		{name: "no account type policy", allowed: true, wantAccessLookup: true},
		{name: "no active grant", owner: &otherID, wantKind: apperr.ACCESS_DENIED, wantAccessLookup: true},
		{name: "professional without grant", accountType: accountdomain.AccountTypeProfessional, wantKind: apperr.ACCESS_DENIED, wantAccessLookup: true},
		{name: "missing patient", missing: true, allowed: true, wantKind: apperr.ACCESS_DENIED},
		{name: "patient lookup failure", patientErr: dbErr, wantKind: apperr.INFRA_DATABASE_ERROR},
		{name: "access lookup failure", accessErr: dbErr, wantKind: apperr.INFRA_DATABASE_ERROR, wantAccessLookup: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patients := &patientLookup{result: &profiledomain.Patient{ID: patientID, OwnerUserID: tc.owner}, err: tc.patientErr}
			if tc.missing {
				patients.result = nil
			}
			access := &accessLookup{allowed: tc.allowed, err: tc.accessErr}
			err := New(patients, access).RequirePatientAccess(context.Background(), &accountdomain.User{ID: actorID, AccountType: tc.accountType}, patientID)
			if tc.wantKind == "" {
				if err != nil {
					t.Fatalf("unexpected denial: %v", err)
				}
			} else {
				var appErr *apperr.AppError
				if !errors.As(err, &appErr) || appErr.Kind != tc.wantKind {
					t.Fatalf("expected %s, got %v", tc.wantKind, err)
				}
				if (tc.patientErr != nil || tc.accessErr != nil) && !errors.Is(err, dbErr) {
					t.Fatalf("database cause was lost: %v", err)
				}
			}
			if patients.id != patientID {
				t.Fatal("wrong patient queried")
			}
			if tc.wantAccessLookup {
				if access.patientID != patientID || access.actorID != actorID {
					t.Fatal("grant checked for the wrong patient or user")
				}
			} else if access.actorID != uuid.Nil {
				t.Fatal("unexpected grant lookup")
			}
		})
	}
}

func TestRequirePatientAccessRejectsMissingIdentityAndConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name      string
		actor     *accountdomain.User
		patientID uuid.UUID
		kind      apperr.ErrorKind
	}{
		{name: "missing user", patientID: uuid.New(), kind: apperr.AUTH_REQUIRED},
		{name: "empty user ID", actor: &accountdomain.User{}, patientID: uuid.New(), kind: apperr.AUTH_REQUIRED},
		{name: "empty patient ID", actor: &accountdomain.User{ID: uuid.New()}, kind: apperr.INTERNAL_ERROR},
		{name: "missing repositories", actor: &accountdomain.User{ID: uuid.New()}, patientID: uuid.New(), kind: apperr.INTERNAL_ERROR},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := New(nil, nil).RequirePatientAccess(context.Background(), tc.actor, tc.patientID)
			var appErr *apperr.AppError
			if !errors.As(err, &appErr) || appErr.Kind != tc.kind {
				t.Fatalf("expected %s, got %v", tc.kind, err)
			}
		})
	}
}
