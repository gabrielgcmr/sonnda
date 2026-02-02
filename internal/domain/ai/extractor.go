// internal/domain/ai/extractor.go
package ai

import "context"

// DocumentExtractorService define a extração de laudos via provider externo.
type DocumentExtractorService interface {
	ExtractLabReport(ctx context.Context, documentURI, mimeType string) (*ExtractedLabReport, error)
}
