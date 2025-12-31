// internal/app/services/patient/service.go
package patientsvc

import (
	"context"

	"sonnda-api/internal/domain/entities/patient"
	"sonnda-api/internal/domain/entities/user"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, currentUser *user.User, input CreateInput) (*patient.Patient, error)
	GetByID(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error)
	UpdateByID(ctx context.Context, currentUser *user.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error)
	SoftDeleteByID(ctx context.Context, currentUser *user.User, id uuid.UUID) error
	List(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error)
	ListByName(ctx context.Context, currentUser *user.User, name string, limit, offset int) ([]*patient.Patient, error)
}
