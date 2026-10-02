// internal/features/documentprocessing/extraction/service_test.go
package extraction

import (
	"context"
	"errors"
	"strings"
	"testing"

	lab "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"
	text "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type readerStub struct{ err error }

func (r readerStub) Extract(context.Context, text.ExtractInput) (*text.ExtractOutput, error) {
	return &text.ExtractOutput{Text: "Glicose 90"}, r.err
}

type labStub struct {
	report *lab.ExtractedLabReport
	err    error
}

func (s labStub) ExtractLabReport(context.Context, lab.ExtractLabReportInput) (*lab.ExtractedLabReport, error) {
	return s.report, s.err
}

func sample() *lab.ExtractedLabReport {
	value, unit, raw, date := "< 5", "mg/dL", "Glicose < 5", "2026-09-01"
	return &lab.ExtractedLabReport{Metadata: lab.ExtractionMetadata{Provider: "test"}, Tests: []lab.ExtractedTestResult{{TestName: "Glicose", CollectedAt: &date, RawText: &raw, Items: []lab.ExtractedTestItem{{ParameterName: "Glicose", ResultValue: &value, ResultUnit: &unit, RawText: &raw, Status: lab.ExtractionStatusPartial, Warnings: []lab.ExtractionWarning{{Code: "test", Message: "Conferir"}}}}}}}
}
func TestExtractionPreservesResultsAndPrivateMetadata(t *testing.T) {
	service := New(readerStub{}, labStub{report: sample()})
	result, err := service.ExtractPDF(context.Background(), "test.pdf", "test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.SummaryText, "Glicose: < 5 mg/dL") || !Usable(&result.Report) {
		t.Fatalf("bad summary: %+v", result)
	}
	if result.Report.Metadata.Provider != "test" {
		t.Fatalf("metadata lost: %+v", result.Report.Metadata)
	}
	test := result.Report.Tests[0]
	if test.RawText == nil || *test.RawText != "Glicose < 5" {
		t.Fatalf("test raw text lost: %+v", test)
	}
	if len(test.Items) != 1 || test.Items[0].Status != lab.ExtractionStatusPartial {
		t.Fatalf("item status lost: %+v", test.Items)
	}
}
func TestExtractionDistinguishesUnreadablePDFAndTechnicalFailure(t *testing.T) {
	for _, tc := range []struct {
		err  error
		kind apperr.ErrorKind
	}{{text.ErrUnreadablePDF, apperr.DOMAIN_RULE_VIOLATION}, {errors.New("pdftotext missing"), apperr.INFRA_EXTERNAL_SERVICE_ERROR}, {context.DeadlineExceeded, apperr.INFRA_TIMEOUT}} {
		_, err := New(readerStub{tc.err}, labStub{}).ExtractPDF(context.Background(), "x.pdf", "x.pdf")
		var app *apperr.AppError
		if !errors.As(err, &app) || app.Kind != tc.kind {
			t.Fatalf("wrong error: %v", err)
		}
	}
}
func TestPartialAndEmptyExtraction(t *testing.T) {
	report := sample()
	report.Tests[0].Items = append(report.Tests[0].Items, lab.ExtractedTestItem{ParameterName: "Outro"})
	invalid := "impossible"
	report.Tests[0].CollectedAt = &invalid
	result, err := New(nil, labStub{report: report}).ExtractText(context.Background(), "laudo")
	if err != nil || result.Status != lab.ExtractionStatusPartial || len(result.Warnings) == 0 || result.Report.Tests[0].CollectedAt != nil {
		t.Fatalf("invalid partial result: %+v %v", result, err)
	}
	result, err = New(nil, labStub{report: &lab.ExtractedLabReport{}}).ExtractText(context.Background(), "pedido")
	if err != nil || Usable(&result.Report) || result.Status != lab.ExtractionStatusNeedsReview {
		t.Fatalf("empty report accepted: %+v", result)
	}
}
