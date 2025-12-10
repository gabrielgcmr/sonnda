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
	ResultText    *string
	ResultUnit    *string
	ReferenceText *string
}
