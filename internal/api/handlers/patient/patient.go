// internal/api/handlers/patient/patient.go
package patient

import (
	"context"

	base "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	patientsvc "github.com/gabrielgcmr/sonnda/internal/application/services/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, currentUser *user.User, input patientsvc.CreateInput) (*patient.Patient, error)
	Get(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error)
	Update(ctx context.Context, currentUser *user.User, id uuid.UUID, input patientsvc.UpdateInput) (*patient.Patient, error)
	HardDelete(ctx context.Context, currentUser *user.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error)
}

type Handler = base.PatientHandler

func NewHandler(svc Service) *Handler {
	return base.NewPatientHandler(svc)
}
