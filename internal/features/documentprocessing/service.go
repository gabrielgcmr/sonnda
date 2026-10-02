// internal/features/documentprocessing/service.go
package documentprocessing

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ExamDocumentOutput, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentOutput, error)
	ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentTextOutput, error)
}
