package ai

import "context"

type DocumentExtractorService interface {
	ExtractLabReport(ctx context.Context, documentURI, mimeType string) (*ExtractedLabReport, error)
}
