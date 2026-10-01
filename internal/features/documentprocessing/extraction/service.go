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
	report, err := s.lab.ExtractLabReport(ctx, labextraction.ExtractLabReportInput{Text: input})
	if err != nil {
		return nil, technicalError("Não foi possível extrair os dados laboratoriais.", err)
	}
	if report == nil {
		return nil, apperr.Internal("O extrator não retornou um resultado.", nil)
	}
	report.Normalize()
	report.RawText = &input
	normalizeDates(report)
	normalizeResults(report)
	status := report.Metadata.Status
	if status == "" {
		status = labextraction.ExtractionStatusSucceeded
	}
	if !Usable(report) {
		status = labextraction.ExtractionStatusNeedsReview
		report.Metadata.Warnings = append(report.Metadata.Warnings, labextraction.ExtractionWarning{Code: "no_usable_results", Message: "Nenhum resultado laboratorial utilizável foi encontrado."})
	} else {
		for _, test := range report.Tests {
			for _, item := range test.Items {
				if item.ResultValue == nil {
					report.Metadata.Warnings = append(report.Metadata.Warnings, labextraction.ExtractionWarning{Code: "missing_value", Message: "Há parâmetros sem valor identificado.", Field: "tests.items.result_value"})
					break
				}
			}
		}
		if len(report.Metadata.Warnings) > 0 && status == labextraction.ExtractionStatusSucceeded {
			status = labextraction.ExtractionStatusPartial
		}
	}
	report.Metadata.Status = status
	warnings := report.Metadata.Warnings
	if warnings == nil {
		warnings = []labextraction.ExtractionWarning{}
	}
	return &Result{Status: status, Warnings: warnings, Report: *report, SummaryText: FormatSummary(report)}, nil
}

func Usable(report *labextraction.ExtractedLabReport) bool {
	if report == nil || report.Metadata.Status == labextraction.ExtractionStatusFailed {
		return false
	}
	for _, test := range report.Tests {
		if strings.TrimSpace(test.TestName) == "" {
			continue
		}
		for _, item := range test.Items {
			if strings.TrimSpace(item.ParameterName) != "" && item.ResultValue != nil && strings.TrimSpace(*item.ResultValue) != "" {
				return true
			}
		}
	}
	return false
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
