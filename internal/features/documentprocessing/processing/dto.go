// internal/features/documentprocessing/processing/dto.go
package processing

import (
	"time"

	"github.com/google/uuid"
)

type CreateLabReportFromDocumentInput struct {
	PatientID        uuid.UUID
	ExamDocumentID   *uuid.UUID
	DocumentURI      string
	MimeType         string
	UploadedByUserID uuid.UUID
	CollectionDate   *time.Time
}
