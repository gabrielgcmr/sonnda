// internal/application/services/exams/service.go
package examsvc

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, input CreateExamDocumentInput) (*ExamDocumentOutput, error)
	FindByID(ctx context.Context, id uuid.UUID) (*ExamDocumentOutput, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentOutput, error)
	RouteDocument(ctx context.Context, input RouteExamDocumentInput) (*ExamDocumentOutput, error)
	MarkFailed(ctx context.Context, input MarkExamDocumentFailedInput) (*ExamDocumentOutput, error)
	CreateDocumentTextFromText(ctx context.Context, input CreateExamDocumentTextFromTextInput) (*ExamDocumentTextOutput, error)
	ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentTextOutput, error)
}
