// internal/application/services/labs/dto.go
package labsvc

import (
	"time"

	"github.com/google/uuid"
)

type LabReportOutput struct {
	ID                uuid.UUID          `json:"id"`
	PatientID         uuid.UUID          `json:"patient_id"`
	PatientName       *string            `json:"patient_name,omitempty"`
	PatientDOB        *time.Time         `json:"patient_dob,omitempty"`
	LabName           *string            `json:"lab_name,omitempty"`
	LabPhone          *string            `json:"lab_phone,omitempty"`
	InsuranceProvider *string            `json:"insurance_provider,omitempty"`
	RequestingDoctor  *string            `json:"requesting_doctor,omitempty"`
	TechnicalManager  *string            `json:"technical_manager,omitempty"`
	ReportDate        *time.Time         `json:"report_date,omitempty"`
	UploadedByUserID  uuid.UUID          `json:"uploaded_by_user_id"`
	Fingerprint       *string            `json:"fingerprint,omitempty"`
	TestResults       []TestResultOutput `json:"test_results"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type TestResultOutput struct {
	ID          uuid.UUID        `json:"id"`
	TestName    string           `json:"test_name"`
	Material    *string          `json:"material,omitempty"`
	Method      *string          `json:"method,omitempty"`
	CollectedAt *time.Time       `json:"collected_at,omitempty"`
	ReleaseAt   *time.Time       `json:"release_at,omitempty"`
	Items       []TestItemOutput `json:"items"`
}

type TestItemOutput struct {
	ID            uuid.UUID `json:"id"`
	ParameterName string    `json:"parameter_name"`
	ResultValue   *string   `json:"result_value,omitempty"`
	ResultUnit    *string   `json:"result_unit,omitempty"`
	ReferenceText *string   `json:"reference_text,omitempty"`
}

// Usado em: GET /patients/:patientID/labs/summary.
type LabReportSummaryOutput struct {
	ID           uuid.UUID                `json:"id"`
	PatientID    uuid.UUID                `json:"patient_id"`
	ReportDate   *time.Time               `json:"report_date,omitempty"`
	SummaryTests []LabResultSummaryOutput `json:"summary_tests"`
}

type LabResultSummaryOutput struct {
	TestName    string                    `json:"test_name"`
	CollectedAt *time.Time                `json:"collected_at,omitempty"`
	Items       []ResultItemSummaryOutput `json:"key_results"`
}

type ResultItemSummaryOutput struct {
	ParameterName string  `json:"parameter_name"`
	ResultValue   *string `json:"result_value,omitempty"`
	ResultUnit    *string `json:"result_unit,omitempty"`
}
