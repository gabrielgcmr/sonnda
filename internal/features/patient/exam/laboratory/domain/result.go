// internal/features/patient/exam/laboratory/domain/result.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type LabResult struct {
	ID          uuid.UUID `json:"id"`
	LabReportID uuid.UUID `json:"lab_report_id"`

	TestName string  `json:"test_name"`
	Material *string `json:"material,omitempty"`
	Method   *string `json:"method,omitempty"`

	CollectedAt *time.Time `json:"collected_at,omitempty"`
	ReleaseAt   *time.Time `json:"release_at,omitempty"`

	Items []LabResultItem `json:"items"`
}

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

func (r *LabResult) Normalize() {
	if r == nil {
		return
	}

	r.Material = trimToNil(r.Material)
	r.Method = trimToNil(r.Method)
	r.CollectedAt = utcOrNil(r.CollectedAt)
	r.ReleaseAt = utcOrNil(r.ReleaseAt)
}
