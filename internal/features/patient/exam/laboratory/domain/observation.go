// internal/features/patient/exam/laboratory/domain/observation.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Observation struct {
	ID         uuid.UUID `json:"id"`
	LabPanelID uuid.UUID `json:"lab_panel_id"`

	ParameterName string  `json:"parameter_name"`
	ResultValue   *string `json:"result_value,omitempty"`
	ResultUnit    *string `json:"result_unit,omitempty"`
	ReferenceText *string `json:"reference_text,omitempty"`
}

func NewObservation(labPanelID, parameterName string) (*Observation, error) {
	labPanelID = strings.TrimSpace(labPanelID)
	parameterName = strings.TrimSpace(parameterName)
	if labPanelID == "" {
		return nil, ErrMissingId
	}
	if parameterName == "" {
		return nil, ErrInvalidParameterName
	}

	parsedLabPanelID, err := uuid.Parse(labPanelID)
	if err != nil {
		return nil, ErrMissingId
	}

	return &Observation{
		ID:            uuid.Must(uuid.NewV7()),
		LabPanelID:    parsedLabPanelID,
		ParameterName: parameterName,
	}, nil
}

func (i *Observation) Normalize() {
	if i == nil {
		return
	}

	i.ResultValue = trimToNil(i.ResultValue)
	i.ResultUnit = trimToNil(i.ResultUnit)
	i.ReferenceText = trimToNil(i.ReferenceText)
}

type ObservationTimeline struct {
	ReportID      uuid.UUID  `json:"report_id"`
	LabPanelID    uuid.UUID  `json:"lab_panel_id"`
	ObservationID uuid.UUID  `json:"observation_id"`
	ReportDate    *time.Time `json:"report_date,omitempty"`
	TestName      string     `json:"test_name"`
	ParameterName string     `json:"parameter_name"`
	ResultValue   *string    `json:"result_value,omitempty"`
	ResultUnit    *string    `json:"result_unit,omitempty"`
	ReferenceText *string    `json:"reference_text,omitempty"`
}
