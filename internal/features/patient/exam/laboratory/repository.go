// internal/features/patient/exam/laboratory/repository.go
package laboratory

import (
	"context"

	labdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, report *labdomain.LabReport) error
	FindByID(ctx context.Context, reportID uuid.UUID) (*labdomain.LabReport, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListLabs(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]labdomain.LabReport, error)
	ListItemsByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit, offset int) ([]labdomain.LabResultItemTimeline, error)
}
