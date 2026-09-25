// internal/application/services/patient/access_test.go
package patientsvc

import (
	"context"
	"errors"
	"testing"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type deniedPatientAccess struct {
	err       error
	actorID   uuid.UUID
	patientID uuid.UUID
}

func (a *deniedPatientAccess) RequirePatientAccess(_ context.Context, actor *accountdomain.User, patientID uuid.UUID) error {
	a.actorID, a.patientID = actor.ID, patientID
	return a.err
}

func TestPatientOperationsStopWhenAccessIsDenied(t *testing.T) {
	actor := &accountdomain.User{ID: uuid.New()}
	patientID := uuid.New()
	denied := apperr.Forbidden("acesso negado")
	for _, operation := range []string{"get", "update", "soft delete", "hard delete"} {
		t.Run(operation, func(t *testing.T) {
			authz := &deniedPatientAccess{err: denied}
			// Nil repositories ensure no read or write happens after denial.
			svc := New(nil, nil, authz)
			var err error
			switch operation {
			case "get":
				_, err = svc.Get(context.Background(), actor, patientID)
			case "update":
				_, err = svc.Update(context.Background(), actor, patientID, UpdateInput{})
			case "soft delete":
				err = svc.SoftDelete(context.Background(), actor, patientID)
			case "hard delete":
				err = svc.HardDelete(context.Background(), actor, patientID)
			}
			if !errors.Is(err, denied) {
				t.Fatalf("expected access denial, got %v", err)
			}
			if authz.actorID != actor.ID || authz.patientID != patientID {
				t.Fatal("wrong access check")
			}
		})
	}
}
