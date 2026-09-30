// internal/features/documentprocessing/extraction/summary_test.go
package extraction

import (
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
)

func TestFormatLabSummary(t *testing.T) {
	collectedAt := "2026-05-05 11:48:00"
	labName := "Laboratório Excelência"
	value := "4,40"
	unit := "milhões/mm3"
	report := &labextraction.ExtractedLabReport{
		LabName: &labName,
		Tests: []labextraction.ExtractedTestResult{{
			TestName:    "Hemácias",
			CollectedAt: &collectedAt,
			Items: []labextraction.ExtractedTestItem{{
				ParameterName: "Resultado",
				ResultValue:   &value,
				ResultUnit:    &unit,
			}},
		}},
	}

	got := FormatSummary(report)
	want := "Laboratório: Laboratório Excelência\nColeta: 05/05/2026 11:48\n\nHemácias\nResultado: 4,40 milhões/mm3\n"
	if got != want {
		t.Fatalf("FormatSummary() = %q, want %q", got, want)
	}
	if strings.Contains(got, "referência") || strings.Contains(got, "paciente") {
		t.Fatalf("summary contains fields that should be omitted: %q", got)
	}
}

func TestFormatLabSummaryKeepsTextResultsAndMultipleDates(t *testing.T) {
	firstDate := "2026-05-05"
	secondDate := "2026-05-06"
	negative := "Negativo"
	report := &labextraction.ExtractedLabReport{
		Tests: []labextraction.ExtractedTestResult{
			{TestName: "Sangue oculto", CollectedAt: &firstDate, Items: []labextraction.ExtractedTestItem{{ParameterName: "Resultado", ResultValue: &negative}}},
			{TestName: "Outro exame", CollectedAt: &secondDate},
		},
	}

	got := FormatSummary(report)
	if !strings.Contains(got, "Coleta: 05/05/2026\nResultado: Negativo") || !strings.Contains(got, "Coleta: 06/05/2026") {
		t.Fatalf("FormatSummary() did not preserve text result or per-test dates: %q", got)
	}
}
