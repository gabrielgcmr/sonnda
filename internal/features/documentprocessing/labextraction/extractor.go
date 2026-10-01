// internal/features/documentprocessing/labextraction/extractor.go
package labextraction

import (
	"context"
	"errors"
	"strings"
)

var ErrEmptyText = errors.New("lab extraction requires non-empty text")

// ExtractLabReportInput receives semantic text prepared by the application.
// The application retains the original text, patient and file context.
type ExtractLabReportInput struct {
	Text string
}

func (input ExtractLabReportInput) Validate() error {
	if strings.TrimSpace(input.Text) == "" {
		return ErrEmptyText
	}
	return nil
}

// LabReportTextExtractor decodes structured data and reports provider metadata.
// Normalization and result assessment belong to the extraction application service.
type LabReportTextExtractor interface {
	ExtractLabReport(ctx context.Context, input ExtractLabReportInput) (*ExtractedLabReport, error)
}
