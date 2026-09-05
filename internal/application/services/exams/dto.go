// internal/application/services/exams/dto.go
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

type RouteExamDocumentInput struct {
	ID               uuid.UUID
	ExtractedText    string
	ExtractionMethod string
}

type MarkExamDocumentFailedInput struct {
	ID           uuid.UUID
	ErrorMessage string
}

type CreateExamDocumentTextFromTextInput struct {
	ExamDocumentID   uuid.UUID
	PatientID        uuid.UUID
	UploadedByUserID uuid.UUID
	Category         exams.ExamType
	Text             string
	ExtractionMethod string
	Confidence       *float64
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

type ExamDocumentTextOutput struct {
	ID                 uuid.UUID      `json:"id"`
	ExamDocumentID     *uuid.UUID     `json:"exam_document_id,omitempty"`
	PatientID          uuid.UUID      `json:"patient_id"`
	UploadedByUserID   uuid.UUID      `json:"uploaded_by_user_id"`
	Category           exams.ExamType `json:"category"`
	Title              *string        `json:"title,omitempty"`
	Modality           *string        `json:"modality,omitempty"`
	BodySite           *string        `json:"body_site,omitempty"`
	PerformedAt        *time.Time     `json:"performed_at,omitempty"`
	FacilityName       *string        `json:"facility_name,omitempty"`
	InterpretingDoctor *string        `json:"interpreting_doctor,omitempty"`
	Text               string         `json:"text"`
	Conclusion         *string        `json:"conclusion,omitempty"`
	ExtractionMethod   *string        `json:"extraction_method,omitempty"`
	Confidence         *float64       `json:"confidence,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}
