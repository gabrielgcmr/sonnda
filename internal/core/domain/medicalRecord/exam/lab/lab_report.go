package lab

import "time"

type LabReport struct {
	ID        string
	PatientID string

	PatientName       *string
	PatientDOB        *time.Time
	LabName           *string
	LabPhone          *string
	InsuranceProvider *string
	RequestingDoctor  *string
	TechnicalManager  *string
	ReportDate        *time.Time
	Fingerprint       *string

	RawText *string

	TestResults []LabResult

	CreatedAt        time.Time
	UpdatedAt        time.Time
	UploadedByUserID string
}

type LabResult struct {
	ID          string
	LabReportID string

	TestName string
	Material *string
	Method   *string

	CollectedAt *time.Time
	ReleaseAt   *time.Time

	Items []LabResultItem
}

type LabResultItem struct {
	ID          string
	LabResultID string

	ParameterName string
	ResultValue   *string
	ResultUnit    *string
	ReferenceText *string
}

type LabResultItemTimeline struct {
	ReportID      string
	LabResultID   string
	ItemID        string
	ReportDate    *time.Time
	TestName      string
	ParameterName string
	ResultValue   *string
	ResultUnit    *string
}
