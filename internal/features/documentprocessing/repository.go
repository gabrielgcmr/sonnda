// internal/features/documentprocessing/repository.go
package documentprocessing

import (
	"context"

	processingdomain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	"github.com/google/uuid"
)

type LaboratoryReportRepository interface {
	CreateFromDocument(ctx context.Context, report *laboratory.LabReport, metadata processingdomain.LaboratoryReportMetadata) error
	FindByFingerprint(ctx context.Context, patientID uuid.UUID, fingerprint string) (*laboratory.LabReport, error)
	FindByID(ctx context.Context, reportID uuid.UUID) (*laboratory.LabReport, error)
	AttachDocument(ctx context.Context, reportID, patientID, documentID uuid.UUID) error
}

type DocumentRepository interface {
	Create(ctx context.Context, document *processingdomain.ExamDocument) error
	FindByID(ctx context.Context, id uuid.UUID) (*processingdomain.ExamDocument, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]processingdomain.ExamDocument, error)
	MarkProcessing(ctx context.Context, id uuid.UUID, extractionMethod *string) (*processingdomain.ExamDocument, error)
	MarkClassified(
		ctx context.Context,
		id uuid.UUID,
		status processingdomain.DocumentStatus,
		examType processingdomain.ExamType,
		extractionMethod *string,
		confidence *float64,
		extractedText *string,
		errorMessage *string,
	) (*processingdomain.ExamDocument, error)
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) (*processingdomain.ExamDocument, error)
	CreateDocumentText(ctx context.Context, documentText *processingdomain.ExamDocumentText) error
	FindDocumentTextByDocumentID(ctx context.Context, documentID uuid.UUID) (*processingdomain.ExamDocumentText, error)
	ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]processingdomain.ExamDocumentText, error)
}
