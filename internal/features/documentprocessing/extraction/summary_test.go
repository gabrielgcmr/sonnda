// internal/features/documentprocessing/extraction/summary_test.go
package extraction

import (
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"
)

func TestFormatLabSummary(t *testing.T) {
	collectedAt := "2026-05-05 11:48:00"
	labName := "Laboratório Excelência"
	patientName := "HELLEN CRISTINALLEN CRISTINA"
	value := "4,40"
	unit := "milhões/mm3"
	report := &labextraction.ExtractedLabReport{
		PatientName: &patientName,
		LabName:     &labName,
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
	want := "Paciente: HELLEN CRISTINALLEN CRISTINA\nLaboratório: Laboratório Excelência\nData de coleta: 05/05/2026 11:48\n\nHemácias\nResultado: 4,40 milhões/mm3\n"
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
	if !strings.HasPrefix(got, "Data de coleta: 05/05/2026, 06/05/2026\n\n") || !strings.Contains(got, "Resultado: Negativo") || strings.Count(got, "Data de coleta:") != 1 {
		t.Fatalf("FormatSummary() did not preserve text result or header dates: %q", got)
	}
}

func TestFormatLabSummaryKeepsCollectionTimesOnlyInHeader(t *testing.T) {
	firstDate := "2026-06-19T08:00:00-03:00"
	secondDate := "2026-06-19T09:30:00-03:00"
	report := &labextraction.ExtractedLabReport{
		Tests: []labextraction.ExtractedTestResult{
			{TestName: "Hemograma", CollectedAt: &firstDate},
			{TestName: "Glicemia", CollectedAt: &secondDate},
			{TestName: "Colesterol", CollectedAt: &firstDate},
		},
	}
	got := FormatSummary(report)
	want := "Data de coleta: 19/06/2026 08:00, 19/06/2026 09:30\n\nHemograma\n\nGlicemia\n\nColesterol\n"
	if got != want {
		t.Fatalf("FormatSummary() = %q, want %q", got, want)
	}
}
