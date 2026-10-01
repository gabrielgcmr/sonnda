// internal/features/documentprocessing/extraction/snapshot_compatibility_test.go
package extraction_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	lab "github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/extraction"
)

func baselinePtr[T any](value T) *T { return &value }

func TestSnapshotV1Compatibility(t *testing.T) {
	// Read a fixed historical shape, rather than constructing the input with Encode.
	data, err := os.ReadFile(filepath.Join("testdata", "snapshot_v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := extraction.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	warnings := []lab.ExtractionWarning{{Code: "review_source", Message: "Conferir documento original.", Field: "tests.items"}}
	want := &extraction.Result{
		Status:      lab.ExtractionStatusPartial,
		Warnings:    warnings,
		SummaryText: "Painel A\nMarcador: < 5 mg/L\nTriagem: Negativo\n\nPainel B\nGlicose: 90 mg/dL\n",
		Report: lab.ExtractedLabReport{
			Metadata: lab.ExtractionMetadata{
				Provider: "gemini", Processor: "synthetic-processor", Model: "baseline-model",
				Status: lab.ExtractionStatusPartial, Confidence: baselinePtr(0.9), Warnings: warnings,
			},
			RawText: baselinePtr("  Texto bruto sintético\n"),
			Tests: []lab.ExtractedTestResult{
				{
					TestName: "Painel A", RawText: baselinePtr("Painel A original"),
					Status: lab.ExtractionStatusPartial, Confidence: baselinePtr(0.8), Warnings: warnings,
					Items: []lab.ExtractedTestItem{
						{
							ParameterName: "Marcador", ResultValue: baselinePtr("< 5"), ResultUnit: baselinePtr("mg/L"),
							RawText: baselinePtr("Marcador < 5"), Status: lab.ExtractionStatusPartial,
							Confidence: baselinePtr(0.7), Warnings: warnings,
						},
						{
							ParameterName: "Triagem", ResultValue: baselinePtr("Negativo"),
							RawText: baselinePtr("Triagem Negativo"), Status: lab.ExtractionStatusSucceeded,
						},
					},
				},
				{
					TestName: "Painel B", RawText: baselinePtr("Painel B original"), Status: lab.ExtractionStatusSucceeded,
					Items: []lab.ExtractedTestItem{{
						ParameterName: "Glicose", ResultValue: baselinePtr("90"), ResultUnit: baselinePtr("mg/dL"),
						RawText: baselinePtr("Glicose 90"), Status: lab.ExtractionStatusSucceeded,
						Confidence: baselinePtr(0.95), Warnings: []lab.ExtractionWarning{},
					}},
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot changed public results or private metadata\ngot: %#v\nwant: %#v", got, want)
	}
	encoded, err := extraction.Encode(got)
	if err != nil {
		t.Fatal(err)
	}
	assertBaselineJSON(t, encoded, data)
}
