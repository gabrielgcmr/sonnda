// internal/features/patient/exam/laboratory/domain/panel.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type LabPanel struct {
	ID          uuid.UUID `json:"id"`
	LabReportID uuid.UUID `json:"lab_report_id"`

	TestName string  `json:"test_name"`
	Material *string `json:"material,omitempty"`
	Method   *string `json:"method,omitempty"`

	CollectedAt *time.Time `json:"collected_at,omitempty"`
	ReleaseAt   *time.Time `json:"release_at,omitempty"`

	Items []Observation `json:"observations"`
}

func NewLabPanel(labReportID, testName string) (*LabPanel, error) {
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

	return &LabPanel{
		ID:          uuid.Must(uuid.NewV7()),
		LabReportID: parsedLabReportID,
		TestName:    testName,
		Items:       make([]Observation, 0),
	}, nil
}

func (r *LabPanel) Normalize() {
	if r == nil {
		return
	}

	r.Material = trimToNil(r.Material)
	r.Method = trimToNil(r.Method)
	r.CollectedAt = utcOrNil(r.CollectedAt)
	r.ReleaseAt = utcOrNil(r.ReleaseAt)
}
