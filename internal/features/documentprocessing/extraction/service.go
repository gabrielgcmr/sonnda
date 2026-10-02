// internal/features/documentprocessing/extraction/service.go
package extraction

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"
	text "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type Result struct {
	Status      labextraction.ExtractionStatus    `json:"status"`
	Warnings    []labextraction.ExtractionWarning `json:"warnings"`
	Report      labextraction.ExtractedLabReport  `json:"report"`
	SummaryText string                            `json:"summary_text"`
}

type Service struct {
	text text.Extractor
	lab  labextraction.LabReportTextExtractor
}

func New(reader text.Extractor, lab labextraction.LabReportTextExtractor) *Service {
	return &Service{text: reader, lab: lab}
}

func unavailable(value any) bool {
	return value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
}

func (s *Service) ExtractPDF(ctx context.Context, path, filename string) (*Result, error) {
	if s == nil || unavailable(s.text) || unavailable(s.lab) {
		return nil, apperr.Internal("A extração laboratorial está indisponível.", nil)
	}
	extracted, err := s.text.Extract(ctx, text.ExtractInput{LocalPath: path, OriginalFilename: filename, MimeType: "application/pdf"})
	if err != nil {
		if errors.Is(err, text.ErrUnreadablePDF) {
			return nil, unreadablePDF()
		}
		return nil, technicalError("Não foi possível ler o PDF.", err)
	}
	if extracted == nil || strings.TrimSpace(extracted.Text) == "" {
		return nil, unreadablePDF()
	}
	return s.ExtractText(ctx, extracted.Text)
}

func (s *Service) ExtractText(ctx context.Context, input string) (*Result, error) {
	if s == nil || unavailable(s.lab) {
		return nil, apperr.Internal("A extração laboratorial está indisponível.", nil)
	}
	if strings.TrimSpace(input) == "" {
		return nil, unreadablePDF()
	}
	semanticText := text.NormalizeForSemanticExtraction(input)
	report, err := s.lab.ExtractLabReport(ctx, labextraction.ExtractLabReportInput{Text: semanticText})
	if err != nil {
		return nil, technicalError("Não foi possível extrair os dados laboratoriais.", err)
	}
	if report == nil {
		return nil, apperr.Internal("O extrator não retornou um resultado.", nil)
	}
	report.Normalize()
	report.RawText = &input
	assessStructure(report)
	normalizeDates(report)
	normalizeResults(report)
	assessResults(report)
	warnings := report.Metadata.Warnings
	if warnings == nil {
		warnings = []labextraction.ExtractionWarning{}
	}
	return &Result{Status: report.Metadata.Status, Warnings: warnings, Report: *report, SummaryText: FormatSummary(report)}, nil
}

func unreadablePDF() error {
	return apperr.DomainRuleViolation("Envie um PDF com texto selecionável e legível. PDFs escaneados não são compatíveis.", apperr.Violation{Field: "file", Reason: "text_not_readable"})
}

func technicalError(message string, err error) error {
	kind := apperr.INFRA_EXTERNAL_SERVICE_ERROR
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		kind = apperr.INFRA_TIMEOUT
	}
	return &apperr.AppError{Kind: kind, Message: message, Cause: err}
}
