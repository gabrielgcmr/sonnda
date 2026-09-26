// internal/features/patient/exam/laboratory/service.go
package laboratory

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]LabReportSummaryOutput, error)
	ListFull(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LabReportOutput, error)
	FindByID(ctx context.Context, reportID uuid.UUID) (*LabReportOutput, error)
}
