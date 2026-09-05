// internal/domain/entity/exams/exam_document_text.go
package exams

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ExamDocumentText struct {
	ID                 uuid.UUID  `json:"id"`
	ExamDocumentID     *uuid.UUID `json:"exam_document_id,omitempty"`
	PatientID          uuid.UUID  `json:"patient_id"`
	UploadedByUserID   uuid.UUID  `json:"uploaded_by_user_id"`
	Category           ExamType   `json:"category"`
	Title              *string    `json:"title,omitempty"`
	Modality           *string    `json:"modality,omitempty"`
	BodySite           *string    `json:"body_site,omitempty"`
	PerformedAt        *time.Time `json:"performed_at,omitempty"`
	FacilityName       *string    `json:"facility_name,omitempty"`
	InterpretingDoctor *string    `json:"interpreting_doctor,omitempty"`
	Text               string     `json:"text"`
	Conclusion         *string    `json:"conclusion,omitempty"`
	ExtractionMethod   *string    `json:"extraction_method,omitempty"`
	Confidence         *float64   `json:"confidence,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func NewExamDocumentText(
	examDocumentID *uuid.UUID,
	patientID uuid.UUID,
	uploadedByUserID uuid.UUID,
	category ExamType,
	text string,
) (*ExamDocumentText, error) {
	if patientID == uuid.Nil {
		return nil, ErrInvalidPatientID
	}
	if uploadedByUserID == uuid.Nil {
		return nil, ErrInvalidUploadedByUserID
	}

	now := time.Now().UTC()
	documentText := &ExamDocumentText{
		ID:               uuid.Must(uuid.NewV7()),
		ExamDocumentID:   examDocumentID,
		PatientID:        patientID,
		UploadedByUserID: uploadedByUserID,
		Category:         category,
		Text:             text,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := documentText.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return documentText, nil
}

func (r *ExamDocumentText) NormalizeAndValidate() error {
	if r == nil {
		return ErrInvalidExamDocumentText
	}

	r.Text = strings.TrimSpace(r.Text)
	r.Title = trimToNil(r.Title)
	r.Modality = trimToNil(r.Modality)
	r.BodySite = trimToNil(r.BodySite)
	r.FacilityName = trimToNil(r.FacilityName)
	r.InterpretingDoctor = trimToNil(r.InterpretingDoctor)
	r.Conclusion = trimToNil(r.Conclusion)
	r.ExtractionMethod = trimToNil(r.ExtractionMethod)

	if r.ExamDocumentID != nil && *r.ExamDocumentID == uuid.Nil {
		r.ExamDocumentID = nil
	}
	if !r.Category.Valid() {
		return ErrInvalidExamType
	}
	if r.Text == "" {
		return ErrInvalidText
	}
	if r.Confidence != nil && (*r.Confidence < 0 || *r.Confidence > 1) {
		return ErrInvalidConfidence
	}

	if r.PerformedAt != nil {
		performedAt := r.PerformedAt.UTC()
		r.PerformedAt = &performedAt
	}
	r.CreatedAt = r.CreatedAt.UTC()
	r.UpdatedAt = r.UpdatedAt.UTC()

	return nil
}
