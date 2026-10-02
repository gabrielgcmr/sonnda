// internal/features/documentprocessing/repository.go
package documentprocessing

import (
	"context"

	domain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	"github.com/google/uuid"
)

type DocumentRepository interface {
	FindByID(context.Context, uuid.UUID) (*domain.ExamDocument, error)
	ListByPatient(context.Context, uuid.UUID, int, int) ([]domain.ExamDocument, error)
	ListDocumentTextsByPatient(context.Context, uuid.UUID, int, int) ([]domain.ExamDocumentText, error)
}
