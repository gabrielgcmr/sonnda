package examsvc

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/google/uuid"
)

type CreateExamDocumentInput struct {
	PatientID        uuid.UUID
	UploadedByUserID uuid.UUID
	StorageURI       string
	OriginalFilename string
	MimeType         string
}

type ExamDocumentOutput struct {
	ID               uuid.UUID            `json:"id"`
	PatientID        uuid.UUID            `json:"patient_id"`
	UploadedByUserID uuid.UUID            `json:"uploaded_by_user_id"`
	StorageURI       string               `json:"storage_uri"`
	OriginalFilename string               `json:"original_filename"`
	MimeType         string               `json:"mime_type"`
	Status           exams.DocumentStatus `json:"status"`
	ExamType         *exams.ExamType      `json:"exam_type,omitempty"`
	ExtractionMethod *string              `json:"extraction_method,omitempty"`
	Confidence       *float64             `json:"confidence,omitempty"`
	ErrorMessage     *string              `json:"error_message,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}
