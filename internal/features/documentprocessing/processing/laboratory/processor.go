// internal/features/documentprocessing/processing/laboratory/processor.go
package laboratory

import (
	"context"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	processing "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	labservice "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
)

type extractedReportCreator interface {
	CreateFromExtracted(ctx context.Context, input processing.CreateLabReportFromDocumentInput, extracted *labextraction.ExtractedLabReport) (*labservice.LabReportOutput, error)
}

type documentTextWriter interface {
	CreateDocumentTextFromText(ctx context.Context, input documents.CreateExamDocumentTextFromTextInput) (*documents.ExamDocumentTextOutput, error)
}

type Processor struct {
	extractor labextraction.LabReportTextExtractor
	creator   extractedReportCreator
	documents documentTextWriter
}

var _ processing.Processor = (*Processor)(nil)

func NewProcessor(
	extractor labextraction.LabReportTextExtractor,
	creator extractedReportCreator,
	documentWriter documentTextWriter,
) *Processor {
	return &Processor{extractor: extractor, creator: creator, documents: documentWriter}
}

func (p *Processor) Kind() processing.DocumentKind {
	return processing.DocumentKindLaboratory
}

func (p *Processor) Process(ctx context.Context, input processing.ProcessingInput) error {
	if input.Document == nil || input.ExtractedText == nil || strings.TrimSpace(input.ExtractedText.Text) == "" {
		return processing.ErrProcessorNeedsReview
	}
	if p.extractor == nil || p.creator == nil {
		return processing.ErrProcessorUnavailable
	}

	extracted, err := p.extractor.ExtractLabReport(ctx, labextraction.ExtractLabReportInput{Text: input.ExtractedText.Text})
	if err != nil {
		return err
	}
	if extracted == nil || !extracted.HasStructuredResults() {
		return processing.ErrProcessorNeedsReview
	}

	report, err := p.creator.CreateFromExtracted(ctx, processing.CreateLabReportFromDocumentInput{
		PatientID:        input.Document.PatientID,
		ExamDocumentID:   &input.Document.ID,
		DocumentURI:      input.Document.StorageURI,
		MimeType:         input.Document.MimeType,
		UploadedByUserID: input.Document.UploadedByUserID,
		CollectionDate:   input.CollectionDate,
	}, extracted)
	if err != nil {
		return err
	}
	if report != nil {
		p.createDocumentText(ctx, input, report, extracted)
	}
	return nil
}

func (p *Processor) createDocumentText(
	ctx context.Context,
	input processing.ProcessingInput,
	report *labservice.LabReportOutput,
	extracted *labextraction.ExtractedLabReport,
) {
	if p.documents == nil {
		return
	}
	text := processing.BuildLaboratoryReportText(report)
	if text == "" {
		return
	}
	method := extracted.Metadata.Provider
	if method == "" {
		method = "laboratory_processor"
	}
	_, _ = p.documents.CreateDocumentTextFromText(ctx, documents.CreateExamDocumentTextFromTextInput{
		ExamDocumentID:   input.Document.ID,
		PatientID:        input.Document.PatientID,
		UploadedByUserID: input.Document.UploadedByUserID,
		Category:         processing.DocumentKindLaboratory,
		Text:             text,
		PerformedAt:      input.CollectionDate,
		ExtractionMethod: method,
		Confidence:       extracted.Metadata.Confidence,
	})
}
