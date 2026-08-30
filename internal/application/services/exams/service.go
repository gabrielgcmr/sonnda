package examsvc

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, input CreateExamDocumentInput) (*ExamDocumentOutput, error)
	FindByID(ctx context.Context, id uuid.UUID) (*ExamDocumentOutput, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentOutput, error)
}
