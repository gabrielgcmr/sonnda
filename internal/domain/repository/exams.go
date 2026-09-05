// internal/domain/repository/exams.go
package repository

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/google/uuid"
)

type Exams interface {
	Create(ctx context.Context, document *exams.ExamDocument) error
	FindByID(ctx context.Context, id uuid.UUID) (*exams.ExamDocument, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]exams.ExamDocument, error)
	MarkProcessing(ctx context.Context, id uuid.UUID, extractionMethod *string) (*exams.ExamDocument, error)
	MarkClassified(
		ctx context.Context,
		id uuid.UUID,
		status exams.DocumentStatus,
		examType exams.ExamType,
		extractionMethod *string,
		confidence *float64,
		extractedText *string,
	) (*exams.ExamDocument, error)
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) (*exams.ExamDocument, error)
	CreateDocumentText(ctx context.Context, documentText *exams.ExamDocumentText) error
	FindDocumentTextByDocumentID(ctx context.Context, documentID uuid.UUID) (*exams.ExamDocumentText, error)
	ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]exams.ExamDocumentText, error)
}
