// internal/features/documentprocessing/processing/processor.go
package processing

import (
	"context"
	"errors"
	"time"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	documentdomain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
)

type DocumentKind = documentdomain.ExamType

const (
	DocumentKindLaboratory DocumentKind = documentdomain.ExamTypeLaboratory
	DocumentKindImaging    DocumentKind = documentdomain.ExamTypeImaging
	DocumentKindUnknown    DocumentKind = documentdomain.ExamTypeUnknown
)

type ProcessingInput struct {
	Document       *documents.ExamDocumentOutput
	ExtractedText  *domaintext.ExtractOutput
	CollectionDate *time.Time
}

type Processor interface {
	Kind() DocumentKind
	Process(context.Context, ProcessingInput) error
}

var ErrProcessorNeedsReview = errors.New("processor could not produce validated clinical data")
var ErrProcessorUnavailable = errors.New("document processor is unavailable")
