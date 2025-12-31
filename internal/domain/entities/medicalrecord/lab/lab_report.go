package lab

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type LabReport struct {
	ID        uuid.UUID
	PatientID uuid.UUID

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

	CreatedAt  time.Time
	UpdatedAt  time.Time
	UploadedBy uuid.UUID
}

// NewLabReport creates a report with generated IDs and UTC timestamps.
// Only patientID is required here; caller fills optional metadata afterward.
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

// Normalize fills optional fields with trimmed values and UTC times.
// Call this before persisting if optional metadata was set after creation.
func (r *LabReport) Normalize() {
	if r == nil {
		return
	}

	// Strings: trim, set nil if empty
	r.PatientName = trimToNil(r.PatientName)
	r.LabName = trimToNil(r.LabName)
	r.LabPhone = trimToNil(r.LabPhone)
	r.InsuranceProvider = trimToNil(r.InsuranceProvider)
	r.RequestingDoctor = trimToNil(r.RequestingDoctor)
	r.TechnicalManager = trimToNil(r.TechnicalManager)
	r.Fingerprint = trimToNil(r.Fingerprint)
	r.RawText = trimToNil(r.RawText)

	// Times: force UTC if set
	r.PatientDOB = utcOrNil(r.PatientDOB)
	r.ReportDate = utcOrNil(r.ReportDate)

	// UpdatedAt should remain UTC
	r.UpdatedAt = r.UpdatedAt.UTC()
}

type LabResult struct {
	ID          uuid.UUID
	LabReportID uuid.UUID

	TestName string
	Material *string
	Method   *string

	CollectedAt *time.Time
	ReleaseAt   *time.Time

	Items []LabResultItem
}

// NewLabResult creates a result with generated ID and required fields.
func NewLabResult(labReportID, testName string) (*LabResult, error) {
	labReportID = strings.TrimSpace(labReportID)
	testName = strings.TrimSpace(testName)

	if labReportID == "" {
		return nil, ErrMissingId
	}
	if testName == "" {
		return nil, ErrInvalidTestName
	}

	parsedLabReportID, err := uuid.Parse(labReportID)
	if err != nil {
		return nil, ErrMissingId
	}

	return &LabResult{
		ID:          uuid.Must(uuid.NewV7()),
		LabReportID: parsedLabReportID,
		TestName:    testName,
		Items:       make([]LabResultItem, 0),
	}, nil
}

// Normalize trims optional strings and forces UTC on timestamps.
func (r *LabResult) Normalize() {
	if r == nil {
		return
	}

	r.Material = trimToNil(r.Material)
	r.Method = trimToNil(r.Method)
	r.CollectedAt = utcOrNil(r.CollectedAt)
	r.ReleaseAt = utcOrNil(r.ReleaseAt)
}

type LabResultItem struct {
	ID          uuid.UUID
	LabResultID uuid.UUID

	ParameterName string
	ResultValue   *string
	ResultUnit    *string
	ReferenceText *string
}

// NewLabResultItem creates an item with generated ID and required parameter name.
func NewLabResultItem(labResultID, parameterName string) (*LabResultItem, error) {
	labResultID = strings.TrimSpace(labResultID)
	parameterName = strings.TrimSpace(parameterName)

	if labResultID == "" {
		return nil, ErrMissingId
	}
	if parameterName == "" {
		return nil, ErrInvalidParameterName
	}

	parsedLabResultID, err := uuid.Parse(labResultID)
	if err != nil {
		return nil, ErrMissingId
	}

	return &LabResultItem{
		ID:            uuid.Must(uuid.NewV7()),
		LabResultID:   parsedLabResultID,
		ParameterName: parameterName,
	}, nil
}

// Normalize trims optional strings.
func (i *LabResultItem) Normalize() {
	if i == nil {
		return
	}

	i.ResultValue = trimToNil(i.ResultValue)
	i.ResultUnit = trimToNil(i.ResultUnit)
	i.ReferenceText = trimToNil(i.ReferenceText)
}

type LabResultItemTimeline struct {
	ReportID      uuid.UUID
	LabResultID   uuid.UUID
	ItemID        uuid.UUID
	ReportDate    *time.Time
	TestName      string
	ParameterName string
	ResultValue   *string
	ResultUnit    *string
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

func utcOrNil(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	ut := t.UTC()
	return &ut
}
