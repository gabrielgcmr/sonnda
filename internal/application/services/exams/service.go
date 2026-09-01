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
	CreateReportFromText(ctx context.Context, input CreateExamReportFromTextInput) (*ExamReportOutput, error)
	ListReportsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamReportOutput, error)
}
