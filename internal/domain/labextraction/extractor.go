// internal/domain/labextraction/extractor.go
package labextraction

import "context"

// LabReportExtractor define a extracao estruturada de laudos laboratoriais.
type LabReportExtractor interface {
	ExtractLabReport(ctx context.Context, documentURI, mimeType string) (*ExtractedLabReport, error)
}
