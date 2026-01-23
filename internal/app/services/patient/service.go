// internal/app/services/patient/service.go
package patientsvc

import (
	"context"

	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, currentUser *user.User, input CreateInput) (*patient.Patient, error)
	Get(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error)
	Update(ctx context.Context, currentUser *user.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error)
	SoftDelete(ctx context.Context, currentUser *user.User, id uuid.UUID) error
	HardDelete(ctx context.Context, currentUser *user.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error)
}
