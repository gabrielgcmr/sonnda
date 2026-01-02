// internal/app/services/patient/access_policy.go
package patientsvc

import (
	"context"

	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type AccessPolicy interface {
	CanCreate(ctx context.Context, currentUser *user.User, input CreateInput) error
	CanRead(ctx context.Context, currentUser *user.User, patientID uuid.UUID) error
	CanUpdate(ctx context.Context, currentUser *user.User, patientID uuid.UUID) error
	CanDelete(ctx context.Context, currentUser *user.User, patientID uuid.UUID) error
	CanList(ctx context.Context, currentUser *user.User) error
}
