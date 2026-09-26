// internal/features/documentprocessing/processing/process_stored_document_test.go
package processing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	documentdomain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	"github.com/google/uuid"
)

type processDocumentServiceStub struct {
	documents.Service
	document    *documents.ExamDocumentOutput
	routeKind   documentdomain.ExamType
	routeCalls  []documents.RouteExamDocumentInput
	textCreated bool
	failed      bool
	failedInput documents.MarkExamDocumentFailedInput
}

func (s *processDocumentServiceStub) Create(_ context.Context, input documents.CreateExamDocumentInput) (*documents.ExamDocumentOutput, error) {
	s.document = &documents.ExamDocumentOutput{
		ID: uuid.New(), PatientID: input.PatientID, UploadedByUserID: input.UploadedByUserID,
		StorageURI: input.StorageURI, OriginalFilename: input.OriginalFilename, MimeType: input.MimeType,
	}
	return s.document, nil
}

func (s *processDocumentServiceStub) RouteDocument(_ context.Context, input documents.RouteExamDocumentInput) (*documents.ExamDocumentOutput, error) {
	s.routeCalls = append(s.routeCalls, input)
	kind := s.routeKind
	if kind == "" {
		kind = documentdomain.ExamTypeImaging
	}
	status := documentdomain.DocumentStatusProcessed
	var errorMessage *string
	if input.ProcessingError != nil {
		status = documentdomain.DocumentStatusNeedsReview
		message := input.ProcessingError.Message
		errorMessage = &message
	}
	s.document.ExamType = &kind
	s.document.Status = status
	s.document.ErrorMessage = errorMessage
	return s.document, nil
}

func (s *processDocumentServiceStub) MarkFailed(_ context.Context, input documents.MarkExamDocumentFailedInput) (*documents.ExamDocumentOutput, error) {
	s.failed = true
	s.failedInput = input
	return s.document, nil
}

func (s *processDocumentServiceStub) CreateDocumentTextFromText(_ context.Context, _ documents.CreateExamDocumentTextFromTextInput) (*documents.ExamDocumentTextOutput, error) {
	s.textCreated = true
	return nil, nil
}

type processTextExtractorStub struct{}

func (processTextExtractorStub) Extract(context.Context, domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	return &domaintext.ExtractOutput{Text: "Laudo de imagem", Method: "test_ocr"}, nil
}

type processorCallStub struct{ called bool }

func (p *processorCallStub) Kind() DocumentKind { return DocumentKindLaboratory }
func (p *processorCallStub) Process(context.Context, ProcessingInput) error {
	p.called = true
	return nil
}

type processorErrorStub struct {
	err error
}

func (processorErrorStub) Kind() DocumentKind { return DocumentKindLaboratory }
func (p processorErrorStub) Process(context.Context, ProcessingInput) error {
	return p.err
}

func TestProcessorFailureUsesGenericDocumentMessage(t *testing.T) {
	processingErr := errors.New("processor failed")
	documentService := &processDocumentServiceStub{routeKind: documentdomain.ExamTypeLaboratory}
	useCase := NewProcessStoredDocument(documentService, processTextExtractorStub{}, processorErrorStub{err: processingErr})

	_, err := useCase.Execute(context.Background(), ProcessStoredDocumentInput{
		PatientID: uuid.New(), UploadedByUserID: uuid.New(), StorageURI: "gs://bucket/report.pdf",
		OriginalFilename: "report.pdf", MimeType: "application/pdf",
	})
	if !errors.Is(err, processingErr) {
		t.Fatalf("Execute error = %v, want original processor error", err)
	}
	if !documentService.failed || documentService.failedInput.ErrorMessage != "falha no processamento do documento" {
		t.Fatalf("MarkFailed input = %+v, want generic document processing message", documentService.failedInput)
	}
}

func TestUnregisteredDocumentKindMovesToReviewWithoutClinicalProcessor(t *testing.T) {
	documentService := &processDocumentServiceStub{}
	processor := &processorCallStub{}
	useCase := NewProcessStoredDocument(documentService, processTextExtractorStub{}, processor)

	document, err := useCase.Execute(context.Background(), ProcessStoredDocumentInput{
		PatientID: uuid.New(), UploadedByUserID: uuid.New(), StorageURI: "gs://bucket/image.pdf",
		OriginalFilename: "image.pdf", MimeType: "application/pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if processor.called {
		t.Fatal("laboratory processor ran for an imaging document")
	}
	if document.Status != documentdomain.DocumentStatusNeedsReview {
		t.Fatalf("status = %q, want needs_review", document.Status)
	}
	if document.ErrorMessage == nil || !strings.Contains(*document.ErrorMessage, "Nao existe processador") {
		t.Fatalf("missing review reason: %v", document.ErrorMessage)
	}
	if documentService.failed || !documentService.textCreated || len(documentService.routeCalls) != 2 || documentService.routeCalls[1].ProcessingError == nil {
		t.Fatal("unregistered kind must be reviewed, not failed or clinically processed")
	}
}

func TestUnknownDocumentKindCannotRemainProcessed(t *testing.T) {
	documentService := &processDocumentServiceStub{routeKind: documentdomain.ExamTypeUnknown}
	useCase := NewProcessStoredDocument(documentService, processTextExtractorStub{})

	document, err := useCase.Execute(context.Background(), ProcessStoredDocumentInput{
		PatientID: uuid.New(), UploadedByUserID: uuid.New(), StorageURI: "gs://bucket/unknown.pdf",
		OriginalFilename: "unknown.pdf", MimeType: "application/pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if document.Status != documentdomain.DocumentStatusNeedsReview || len(documentService.routeCalls) != 2 {
		t.Fatalf("unknown kind escaped review: status=%q routeCalls=%d", document.Status, len(documentService.routeCalls))
	}
}

func TestProcessorNeedsReviewDoesNotMarkDocumentFailed(t *testing.T) {
	documentService := &processDocumentServiceStub{}
	processor := processorNeedsReviewStub{}
	useCase := NewProcessStoredDocument(documentService, processTextExtractorStub{}, processor)

	document, err := useCase.Execute(context.Background(), ProcessStoredDocumentInput{
		PatientID: uuid.New(), UploadedByUserID: uuid.New(), StorageURI: "gs://bucket/lab.pdf",
		OriginalFilename: "hemograma.pdf", MimeType: "application/pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if document.Status != documentdomain.DocumentStatusNeedsReview || documentService.failed {
		t.Fatalf("empty structured result should be reviewed, got status=%q failed=%v", document.Status, documentService.failed)
	}
}

type processorNeedsReviewStub struct{}

func (processorNeedsReviewStub) Kind() DocumentKind { return DocumentKindLaboratory }
func (processorNeedsReviewStub) Process(context.Context, ProcessingInput) error {
	return ErrProcessorNeedsReview
}

func TestDocumentTextForDisplayUsesNormalizedTextAndKeepsRawFallback(t *testing.T) {
	if got := documentTextForDisplay(&domaintext.ExtractOutput{Text: "Hematocrito 43,8 \uFF05", NormalizedText: "Hematocrito 43,8 %"}); got != "Hematocrito 43,8 %" {
		t.Fatalf("normalized display text = %q", got)
	}
	if got := documentTextForDisplay(&domaintext.ExtractOutput{Text: "Hematocrito 43,8 \uFF05"}); got != "Hematocrito 43,8 \uFF05" {
		t.Fatalf("raw fallback text = %q", got)
	}
}

func TestBuildLaboratoryReportTextFormatsStructuredResults(t *testing.T) {
	reportDate := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	patientName := "Gabriel Cactus Moreno Reboucas"
	labName := "Laboratorio Exemplo"
	unit := "g/dL"
	value := "15,1"
	reference := "13,5 a 17,5"

	text := BuildLaboratoryReportText(&laboratory.LabReportOutput{
		PatientName: &patientName,
		LabName:     &labName,
		ReportDate:  &reportDate,
		TestResults: []laboratory.TestResultOutput{{
			TestName: "HEMOGRAMA",
			Items: []laboratory.TestItemOutput{{
				ParameterName: "Hemoglobina",
				ResultValue:   &value,
				ResultUnit:    &unit,
				ReferenceText: &reference,
			}},
		}},
	})

	for _, part := range []string{
		"Exame laboratorial",
		"Paciente: Gabriel Cactus Moreno Reboucas",
		"Laboratorio: Laboratorio Exemplo",
		"Data do laudo: 31/08/2026",
		"HEMOGRAMA",
		"- Hemoglobina 15,1 g/dL (Referencia: 13,5 a 17,5)",
	} {
		if !strings.Contains(text, part) {
			t.Fatalf("expected text to contain %q, got:\n%s", part, text)
		}
	}
}
