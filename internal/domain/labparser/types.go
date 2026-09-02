package labparser

import "context"

type ParseStatus string

const (
	ParseStatusParsed   ParseStatus = "parsed"
	ParseStatusPartial  ParseStatus = "partial"
	ParseStatusUnparsed ParseStatus = "unparsed"
)

type ParseInput struct {
	RawText      string
	ExamTypeHint string
}

type ParseOutput struct {
	RawText        string
	NormalizedText string
	ExamType       string
	Results        []ParsedLabResult
}

type ParsedLabResult struct {
	Code         string
	OriginalName string
	Value        *float64
	RawValue     string
	Unit         *string
	ReferenceMin *float64
	ReferenceMax *float64
	RawReference *string
	Status       ParseStatus
	RawLine      string
}

type ReferenceRange struct {
	Min  *float64
	Max  *float64
	Text string
}

type Parser interface {
	Parse(ctx context.Context, input ParseInput) (*ParseOutput, error)
}
