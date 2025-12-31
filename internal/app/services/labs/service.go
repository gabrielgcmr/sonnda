package labsvc

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	CreateFromDocument(ctx context.Context, input CreateFromDocumentInput) (*LabReportOutput, error)
	List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]LabReportSummaryOutput, error)
	ListFull(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LabReportOutput, error)
}
