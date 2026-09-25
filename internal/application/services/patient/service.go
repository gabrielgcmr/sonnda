// internal/application/services/patient/service.go
package patientsvc

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, currentUser *accountdomain.User, input CreateInput) (*patient.Patient, error)
	Get(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) (*patient.Patient, error)
	Update(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error)
	SoftDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	HardDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *accountdomain.User, limit, offset int) ([]*patient.Patient, error)
}
