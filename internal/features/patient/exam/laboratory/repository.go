// internal/features/patient/exam/laboratory/repository.go
package laboratory

import (
	"context"

	labdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, reportID uuid.UUID) (*labdomain.LabReport, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListLabs(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]labdomain.LabReport, error)
	ListObservationTimelineByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit, offset int) ([]labdomain.ObservationTimeline, error)
}
