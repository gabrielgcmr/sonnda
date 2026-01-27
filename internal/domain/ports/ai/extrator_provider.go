package ai

import "context"

type ExtractorProvider interface {
	ExtractLabReport(ctx context.Context, documentURI, mimeType string) (*ExtractedLabReport, error)
}
