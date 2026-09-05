// internal/domain/textextraction/extractor.go
package textextraction

import "context"

type ExtractInput struct {
	LocalPath        string
	MimeType         string
	OriginalFilename string
}

type ExtractOutput struct {
	Text   string
	Method string
}

type Extractor interface {
	Extract(ctx context.Context, input ExtractInput) (*ExtractOutput, error)
}
