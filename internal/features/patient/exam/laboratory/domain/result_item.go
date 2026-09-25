// internal/features/patient/exam/laboratory/domain/result_item.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type LabResultItem struct {
	ID          uuid.UUID `json:"id"`
	LabResultID uuid.UUID `json:"lab_result_id"`

	ParameterName string  `json:"parameter_name"`
	ResultValue   *string `json:"result_value,omitempty"`
	ResultUnit    *string `json:"result_unit,omitempty"`
	ReferenceText *string `json:"reference_text,omitempty"`
}

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

func (i *LabResultItem) Normalize() {
	if i == nil {
		return
	}

	i.ResultValue = trimToNil(i.ResultValue)
	i.ResultUnit = trimToNil(i.ResultUnit)
	i.ReferenceText = trimToNil(i.ReferenceText)
}

type LabResultItemTimeline struct {
	ReportID      uuid.UUID  `json:"report_id"`
	LabResultID   uuid.UUID  `json:"lab_result_id"`
	ItemID        uuid.UUID  `json:"item_id"`
	ReportDate    *time.Time `json:"report_date,omitempty"`
	TestName      string     `json:"test_name"`
	ParameterName string     `json:"parameter_name"`
	ResultValue   *string    `json:"result_value,omitempty"`
	ResultUnit    *string    `json:"result_unit,omitempty"`
}
