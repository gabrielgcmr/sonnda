package labsvc

import (
	"time"

	"github.com/google/uuid"
)

type LabReportOutput struct {
	ID                uuid.UUID
	PatientID         uuid.UUID
	PatientName       *string
	PatientDOB        *time.Time
	LabName           *string
	LabPhone          *string
	InsuranceProvider *string
	RequestingDoctor  *string
	TechnicalManager  *string
	ReportDate        *time.Time
	UploadedByUserID  uuid.UUID
	Fingerprint       *string
	TestResults       []TestResultOutput
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TestResultOutput struct {
	ID          uuid.UUID
	TestName    string
	Material    *string
	Method      *string
	CollectedAt *time.Time
	ReleaseAt   *time.Time
	Items       []TestItemOutput
}

type TestItemOutput struct {
	ID            uuid.UUID
	ParameterName string
	ResultValue   *string
	ResultUnit    *string
	ReferenceText *string
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
