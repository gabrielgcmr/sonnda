// internal/features/patient/exam/laboratory/domain/report.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type LabReport struct {
	ID        uuid.UUID `json:"id"`
	PatientID uuid.UUID `json:"patient_id"`

	ExamDocumentID    *uuid.UUID  `json:"exam_document_id,omitempty"`
	PatientName       *string     `json:"patient_name,omitempty"`
	PatientDOB        *time.Time  `json:"patient_dob,omitempty"`
	LabName           *string     `json:"lab_name,omitempty"`
	LabPhone          *string     `json:"lab_phone,omitempty"`
	InsuranceProvider *string     `json:"insurance_provider,omitempty"`
	RequestingDoctor  *string     `json:"requesting_doctor,omitempty"`
	TechnicalManager  *string     `json:"technical_manager,omitempty"`
	ReportDate        *time.Time  `json:"report_date,omitempty"`
	TestResults       []LabResult `json:"test_results"`

	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
}

func NewLabReport(patientID, uploadedBy string) (*LabReport, error) {
	patientID = strings.TrimSpace(patientID)
	uploadedBy = strings.TrimSpace(uploadedBy)

	if patientID == "" {
		return nil, ErrInvalidPatientID
	}
	if uploadedBy == "" {
		return nil, ErrInvalidUploadedByUser
	}

	parsedPatientID, err := uuid.Parse(patientID)
	if err != nil {
		return nil, ErrInvalidPatientID
	}
	parsedUploadedBy, err := uuid.Parse(uploadedBy)
	if err != nil {
		return nil, ErrInvalidUploadedByUser
	}

	now := time.Now().UTC()
	return &LabReport{
		ID:         uuid.Must(uuid.NewV7()),
		PatientID:  parsedPatientID,
		CreatedAt:  now,
		UpdatedAt:  now,
		UploadedBy: parsedUploadedBy,
	}, nil
}

func (r *LabReport) Normalize() {
	if r == nil {
		return
	}

	r.PatientName = trimToNil(r.PatientName)
	r.LabName = trimToNil(r.LabName)
	r.LabPhone = trimToNil(r.LabPhone)
	r.InsuranceProvider = trimToNil(r.InsuranceProvider)
	r.RequestingDoctor = trimToNil(r.RequestingDoctor)
	r.TechnicalManager = trimToNil(r.TechnicalManager)
	r.PatientDOB = utcOrNil(r.PatientDOB)
	r.ReportDate = utcOrNil(r.ReportDate)
	r.UpdatedAt = r.UpdatedAt.UTC()
}

func trimToNil(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func utcOrNil(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
