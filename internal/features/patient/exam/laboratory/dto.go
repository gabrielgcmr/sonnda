// internal/features/patient/exam/laboratory/dto.go
package laboratory

import (
	"time"

	"github.com/google/uuid"
)

type LabReportOutput struct {
	ID                uuid.UUID        `json:"id"`
	PatientID         uuid.UUID        `json:"patient_id"`
	ExamDocumentID    *uuid.UUID       `json:"exam_document_id,omitempty"`
	PatientName       *string          `json:"patient_name,omitempty"`
	PatientDOB        *time.Time       `json:"patient_dob,omitempty"`
	LabName           *string          `json:"lab_name,omitempty"`
	LabPhone          *string          `json:"lab_phone,omitempty"`
	InsuranceProvider *string          `json:"insurance_provider,omitempty"`
	RequestingDoctor  *string          `json:"requesting_doctor,omitempty"`
	TechnicalManager  *string          `json:"technical_manager,omitempty"`
	ReportDate        *time.Time       `json:"report_date,omitempty"`
	UploadedByUserID  uuid.UUID        `json:"uploaded_by_user_id"`
	Panels            []LabPanelOutput `json:"panels"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

type LabPanelOutput struct {
	ID           uuid.UUID           `json:"id"`
	TestName     string              `json:"test_name"`
	Material     *string             `json:"material,omitempty"`
	Method       *string             `json:"method,omitempty"`
	CollectedAt  *time.Time          `json:"collected_at,omitempty"`
	ReleaseAt    *time.Time          `json:"release_at,omitempty"`
	Observations []ObservationOutput `json:"observations"`
}

type ObservationOutput struct {
	ID            uuid.UUID `json:"id"`
	ParameterName string    `json:"parameter_name"`
	ResultValue   *string   `json:"result_value,omitempty"`
	ResultUnit    *string   `json:"result_unit,omitempty"`
	ReferenceText *string   `json:"reference_text,omitempty"`
}

// Usado em: GET /patients/:patientID/labs/summary.
type LabReportSummaryOutput struct {
	ID             uuid.UUID               `json:"id"`
	PatientID      uuid.UUID               `json:"patient_id"`
	ExamDocumentID *uuid.UUID              `json:"exam_document_id,omitempty"`
	ReportDate     *time.Time              `json:"report_date,omitempty"`
	SummaryPanels  []LabPanelSummaryOutput `json:"summary_panels"`
}

type LabPanelSummaryOutput struct {
	TestName     string                     `json:"test_name"`
	CollectedAt  *time.Time                 `json:"collected_at,omitempty"`
	Observations []ObservationSummaryOutput `json:"observations"`
}

type ObservationSummaryOutput struct {
	ParameterName string  `json:"parameter_name"`
	ResultValue   *string `json:"result_value,omitempty"`
	ResultUnit    *string `json:"result_unit,omitempty"`
	ReferenceText *string `json:"reference_text,omitempty"`
}
