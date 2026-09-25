// internal/features/documentprocessing/processing/process_stored_document_test.go
package processing

import (
	"strings"
	"testing"
	"time"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
)

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

	text := buildLaboratoryReportText(&laboratory.LabReportOutput{
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
