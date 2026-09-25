// internal/features/documentprocessing/processing/laboratory/processor_test.go
package laboratory

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	processing "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	labservice "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	"github.com/google/uuid"
)

type textExtractorStub struct {
	input  labextraction.ExtractLabReportInput
	result *labextraction.ExtractedLabReport
	err    error
}

func (s *textExtractorStub) ExtractLabReport(_ context.Context, input labextraction.ExtractLabReportInput) (*labextraction.ExtractedLabReport, error) {
	s.input = input
	return s.result, s.err
}

type reportCreatorStub struct {
	input     processing.CreateLabReportFromDocumentInput
	extracted *labextraction.ExtractedLabReport
	output    *labservice.LabReportOutput
	err       error
	called    bool
}

func (s *reportCreatorStub) CreateFromExtracted(_ context.Context, input processing.CreateLabReportFromDocumentInput, extracted *labextraction.ExtractedLabReport) (*labservice.LabReportOutput, error) {
	s.called = true
	s.input = input
	s.extracted = extracted
	return s.output, s.err
}

type documentTextWriterStub struct {
	input  documents.CreateExamDocumentTextFromTextInput
	called bool
}

func (s *documentTextWriterStub) CreateDocumentTextFromText(_ context.Context, input documents.CreateExamDocumentTextFromTextInput) (*documents.ExamDocumentTextOutput, error) {
	s.called = true
	s.input = input
	return nil, nil
}

func TestProcessorUsesExtractedTextAndPersistsStructuredReport(t *testing.T) {
	patientID, userID, documentID := uuid.New(), uuid.New(), uuid.New()
	collectionDate := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	text := "HEMOGRAMA\nHemoglobina 15,1 g/dL"
	value, unit, name := "15,1", "g/dL", "Hemoglobina"
	extracted := &labextraction.ExtractedLabReport{
		Metadata: labextraction.ExtractionMetadata{Provider: "gemini"},
		Tests: []labextraction.ExtractedTestResult{{
			TestName: "HEMOGRAMA",
			Items:    []labextraction.ExtractedTestItem{{ParameterName: name, ResultValue: &value, ResultUnit: &unit}},
		}},
	}
	extractor := &textExtractorStub{result: extracted}
	creator := &reportCreatorStub{output: &labservice.LabReportOutput{
		PatientID: patientID,
		TestResults: []labservice.TestResultOutput{{
			TestName: "HEMOGRAMA",
			Items:    []labservice.TestItemOutput{{ParameterName: name, ResultValue: &value, ResultUnit: &unit}},
		}},
	}}
	documentWriter := &documentTextWriterStub{}
	processor := NewProcessor(extractor, creator, documentWriter)
	document := &documents.ExamDocumentOutput{
		ID: documentID, PatientID: patientID, UploadedByUserID: userID,
		StorageURI: "gs://bucket/exam.pdf", MimeType: "application/pdf",
	}

	if processor.Kind() != processing.DocumentKindLaboratory {
		t.Fatalf("kind = %q, want laboratory", processor.Kind())
	}
	if err := processor.Process(context.Background(), processing.ProcessingInput{
		Document:       document,
		ExtractedText:  &domaintext.ExtractOutput{Text: text, Method: "document_ai_ocr"},
		CollectionDate: &collectionDate,
	}); err != nil {
		t.Fatal(err)
	}

	if extractor.input.Text != text {
		t.Fatalf("extractor input = %q, want OCR text %q", extractor.input.Text, text)
	}
	if !creator.called || creator.input.PatientID != patientID || creator.input.UploadedByUserID != userID || creator.input.ExamDocumentID == nil || *creator.input.ExamDocumentID != documentID {
		t.Fatal("structured clinical report was not sent to persistence")
	}
	if creator.extracted != extracted {
		t.Fatal("processor did not pass the structured extractor result to the creator")
	}
	if !documentWriter.called || documentWriter.input.Category != processing.DocumentKindLaboratory || documentWriter.input.PerformedAt == nil || !documentWriter.input.PerformedAt.Equal(collectionDate) || !strings.Contains(documentWriter.input.Text, "- Hemoglobina 15,1 g/dL") {
		t.Fatalf("structured document text not preserved: %+v", documentWriter.input)
	}
}

func TestProcessorRequiresReviewWhenNoStructuredResults(t *testing.T) {
	extractor := &textExtractorStub{result: &labextraction.ExtractedLabReport{}}
	creator := &reportCreatorStub{}
	processor := NewProcessor(extractor, creator, &documentTextWriterStub{})
	err := processor.Process(context.Background(), processing.ProcessingInput{
		Document:      &documents.ExamDocumentOutput{ID: uuid.New(), PatientID: uuid.New(), UploadedByUserID: uuid.New()},
		ExtractedText: &domaintext.ExtractOutput{Text: "Atestado medico"},
	})
	if !errors.Is(err, processing.ErrProcessorNeedsReview) {
		t.Fatalf("expected review result, got %v", err)
	}
	if creator.called {
		t.Fatal("clinical persistence must not run when no structured results exist")
	}
}
