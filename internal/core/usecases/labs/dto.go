package labs

import "time"

// CreateFromDocumentInput é o input do use case
type CreateFromDocumentInput struct {
	PatientID        string
	DocumentURI      string
	MimeType         string
	UploadedByUserID string
}

// LabReportOutput é o output do use case
type LabReportOutput struct {
	ID                string
	PatientID         string
	PatientName       *string
	PatientDOB        *time.Time
	LabName           *string
	LabPhone          *string
	InsuranceProvider *string
	RequestingDoctor  *string
	TechnicalManager  *string
	ReportDate        *time.Time
	UploadedByUserID  *string
	TestResults       []TestResultOutput
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TestResultOutput struct {
	ID          string
	TestName    string
	Material    *string
	Method      *string
	CollectedAt *time.Time
	ReleaseAt   *time.Time
	Items       []TestItemOutput
}

type TestItemOutput struct {
	ID            string
	ParameterName string
	ResultValue   *string
	ResultUnit    *string
	ReferenceText *string
}

// Usado em: GET /patients/:patientID/labs
type LabReportSummaryOutput struct {
	ID           string                   `json:"id"`
	PatientID    string                   `json:"patient_id"`
	ReportDate   *time.Time               `json:"report_date,omitempty"`
	SummaryTests []LabResultSummaryOutput `json:"summary_tests"`
}

// Um teste resumido (ex.: "Creatinina", "Hemoglobina")
type LabResultSummaryOutput struct {
	TestName    string `json:"test_name"`
	CollectedAt *time.Time
	Items       []ResultItemSummaryOutput `json:"key_results"`
}

// Resultado essencial de um parâmetro (o que importa pro card/lista)
type ResultItemSummaryOutput struct {
	ParameterName string  `json:"parameter_name"`
	ResultValue   *string `json:"result_value,omitempty"`
	ResultUnit    *string `json:"result_unit,omitempty"`
}
