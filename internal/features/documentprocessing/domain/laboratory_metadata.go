// internal/features/documentprocessing/domain/laboratory_metadata.go
package domain

import (
	"errors"

	"github.com/google/uuid"
)

type LaboratoryReportMetadata struct {
	RawText        *string
	Fingerprint    *string
	ExamDocumentID *uuid.UUID
}

var ErrInvalidDocument = errors.New("invalid document")
