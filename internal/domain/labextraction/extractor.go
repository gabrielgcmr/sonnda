// internal/domain/labextraction/extractor.go
package labextraction

import (
	"context"
	"errors"
	"strings"
)

var ErrEmptyText = errors.New("lab extraction requires non-empty text")

// ExtractLabReportInput recebe texto; paciente e arquivo ficam no caso de uso.
type ExtractLabReportInput struct {
	Text string
}

func (input ExtractLabReportInput) Validate() error {
	if strings.TrimSpace(input.Text) == "" {
		return ErrEmptyText
	}
	return nil
}

// LabReportTextExtractor e o contrato alvo para a LLM da ADR-005.
type LabReportTextExtractor interface {
	ExtractLabReport(ctx context.Context, input ExtractLabReportInput) (*ExtractedLabReport, error)
}
