package exams

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type DocumentStatus string

const (
	DocumentStatusUploaded    DocumentStatus = "uploaded"
	DocumentStatusProcessing  DocumentStatus = "processing"
	DocumentStatusProcessed   DocumentStatus = "processed"
	DocumentStatusFailed      DocumentStatus = "failed"
	DocumentStatusNeedsReview DocumentStatus = "needs_review"
)

type ExamType string

const (
	ExamTypeLaboratory ExamType = "laboratory"
	ExamTypeImaging    ExamType = "imaging"
	ExamTypeUnknown    ExamType = "unknown"
)

type ExamDocument struct {
	ID               uuid.UUID      `json:"id"`
	PatientID        uuid.UUID      `json:"patient_id"`
	UploadedByUserID uuid.UUID      `json:"uploaded_by_user_id"`
	StorageURI       string         `json:"storage_uri"`
	OriginalFilename string         `json:"original_filename"`
	MimeType         string         `json:"mime_type"`
	Status           DocumentStatus `json:"status"`
	ExamType         *ExamType      `json:"exam_type,omitempty"`
	ExtractionMethod *string        `json:"extraction_method,omitempty"`
	Confidence       *float64       `json:"confidence,omitempty"`
	ExtractedText    *string        `json:"extracted_text,omitempty"`
	ErrorMessage     *string        `json:"error_message,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func NewExamDocument(patientID, uploadedByUserID uuid.UUID, storageURI, originalFilename, mimeType string) (*ExamDocument, error) {
	if patientID == uuid.Nil {
		return nil, ErrInvalidPatientID
	}
	if uploadedByUserID == uuid.Nil {
		return nil, ErrInvalidUploadedByUserID
	}

	now := time.Now().UTC()
	document := &ExamDocument{
		ID:               uuid.Must(uuid.NewV7()),
		PatientID:        patientID,
		UploadedByUserID: uploadedByUserID,
		StorageURI:       storageURI,
		OriginalFilename: originalFilename,
		MimeType:         mimeType,
		Status:           DocumentStatusUploaded,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := document.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return document, nil
}

func (d *ExamDocument) NormalizeAndValidate() error {
	if d == nil {
		return ErrInvalidStatus
	}

	d.StorageURI = strings.TrimSpace(d.StorageURI)
	d.OriginalFilename = strings.TrimSpace(d.OriginalFilename)
	d.MimeType = strings.TrimSpace(d.MimeType)
	d.ExtractionMethod = trimToNil(d.ExtractionMethod)
	d.ExtractedText = trimToNil(d.ExtractedText)
	d.ErrorMessage = trimToNil(d.ErrorMessage)

	if d.StorageURI == "" {
		return ErrInvalidStorageURI
	}
	if d.OriginalFilename == "" {
		return ErrInvalidOriginalFilename
	}
	if d.MimeType == "" {
		return ErrInvalidMimeType
	}
	if !d.Status.Valid() {
		return ErrInvalidStatus
	}
	if d.ExamType != nil && !d.ExamType.Valid() {
		return ErrInvalidExamType
	}
	if d.Confidence != nil && (*d.Confidence < 0 || *d.Confidence > 1) {
		return ErrInvalidConfidence
	}

	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()
	return nil
}

func (s DocumentStatus) Valid() bool {
	switch s {
	case DocumentStatusUploaded,
		DocumentStatusProcessing,
		DocumentStatusProcessed,
		DocumentStatusFailed,
		DocumentStatusNeedsReview:
		return true
	default:
		return false
	}
}

func (t ExamType) Valid() bool {
	switch t {
	case ExamTypeLaboratory, ExamTypeImaging, ExamTypeUnknown:
		return true
	default:
		return false
	}
}

func trimToNil(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
