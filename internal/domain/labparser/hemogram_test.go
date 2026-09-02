package labparser

import (
	"context"
	"testing"
)

func TestParseHemogramLine_WithReference(t *testing.T) {
	got, ok := ParseHemogramLine("- Hemoglobina 15,1 g/dL (Referencia: 13,5 a 17,5 g/dL)")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Status != ParseStatusParsed {
		t.Fatalf("expected parsed, got %s", got.Status)
	}
	if got.Code != "hemoglobin" {
		t.Fatalf("expected hemoglobin, got %q", got.Code)
	}
	if got.OriginalName != "Hemoglobina" {
		t.Fatalf("expected original name preserved, got %q", got.OriginalName)
	}
	if got.Value == nil || *got.Value != 15.1 {
		t.Fatalf("expected value 15.1, got %#v", got.Value)
	}
	if got.Unit == nil || *got.Unit != "g/dL" {
		t.Fatalf("expected unit g/dL, got %#v", got.Unit)
	}
	if got.ReferenceMin == nil || *got.ReferenceMin != 13.5 {
		t.Fatalf("expected reference min 13.5, got %#v", got.ReferenceMin)
	}
	if got.ReferenceMax == nil || *got.ReferenceMax != 17.5 {
		t.Fatalf("expected reference max 17.5, got %#v", got.ReferenceMax)
	}
	if got.RawReference == nil || *got.RawReference != "13,5 a 17,5 g/dL" {
		t.Fatalf("expected raw reference preserved, got %#v", got.RawReference)
	}
}

func TestParseHemogramLine_ThousandsValue(t *testing.T) {
	got, ok := ParseHemogramLine("- Plaquetas 453.000 /mm3 (Referencia: 150.000 a 450.000/mm3)")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Code != "platelets" {
		t.Fatalf("expected platelets, got %q", got.Code)
	}
	if got.Value == nil || *got.Value != 453000 {
		t.Fatalf("expected value 453000, got %#v", got.Value)
	}
	if got.ReferenceMin == nil || *got.ReferenceMin != 150000 {
		t.Fatalf("expected reference min 150000, got %#v", got.ReferenceMin)
	}
	if got.ReferenceMax == nil || *got.ReferenceMax != 450000 {
		t.Fatalf("expected reference max 450000, got %#v", got.ReferenceMax)
	}
}

func TestParseHemogramLine_WithoutReference(t *testing.T) {
	got, ok := ParseHemogramLine("- VPM 8,2 /fl")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Status != ParseStatusParsed {
		t.Fatalf("expected parsed, got %s", got.Status)
	}
	if got.Code != "mpv" {
		t.Fatalf("expected mpv, got %q", got.Code)
	}
	if got.RawReference != nil {
		t.Fatalf("expected nil reference, got %#v", got.RawReference)
	}
}

func TestParseHemogramLine_FocusesNameValueAndUnit(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantCode  string
		wantValue float64
		wantUnit  string
	}{
		{
			name:      "rbc",
			line:      "Hemácias 4,98 milhões/ mm³",
			wantCode:  "rbc",
			wantValue: 4.98,
			wantUnit:  "milhões/ mm³",
		},
		{
			name:      "hematocrit",
			line:      "Hematócrito 44,8 %",
			wantCode:  "hematocrit",
			wantValue: 44.8,
			wantUnit:  "%",
		},
		{
			name:      "leukocytes",
			line:      "Leucócitos 8.600 /mm³",
			wantCode:  "leukocytes",
			wantValue: 8600,
			wantUnit:  "/mm³",
		},
		{
			name:      "platelets",
			line:      "Plaquetas 453.000 /mm³",
			wantCode:  "platelets",
			wantValue: 453000,
			wantUnit:  "/mm³",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseHemogramLine(tt.line)
			if !ok {
				t.Fatal("expected line to parse")
			}
			if got.Status != ParseStatusParsed {
				t.Fatalf("expected parsed, got %s", got.Status)
			}
			if got.Code != tt.wantCode {
				t.Fatalf("expected code %q, got %q", tt.wantCode, got.Code)
			}
			if got.Value == nil || *got.Value != tt.wantValue {
				t.Fatalf("expected value %v, got %#v", tt.wantValue, got.Value)
			}
			if got.Unit == nil || *got.Unit != tt.wantUnit {
				t.Fatalf("expected unit %q, got %#v", tt.wantUnit, got.Unit)
			}
		})
	}
}

func TestParseHemogramLine_MalformedReferenceDoesNotBreakNameValue(t *testing.T) {
	got, ok := ParseHemogramLine("- Hemoglobina 15,1 g/dL (Referencia: 13,5 a 17,5 g/dL")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Status != ParseStatusParsed {
		t.Fatalf("expected parsed, got %s", got.Status)
	}
	if got.Code != "hemoglobin" {
		t.Fatalf("expected hemoglobin, got %q", got.Code)
	}
	if got.Value == nil || *got.Value != 15.1 {
		t.Fatalf("expected value 15.1, got %#v", got.Value)
	}
	if got.Unit == nil || *got.Unit != "g/dL" {
		t.Fatalf("expected unit g/dL, got %#v", got.Unit)
	}
}

func TestParseHemogramLine_KnownAnalyteWithoutUnitIsPartial(t *testing.T) {
	got, ok := ParseHemogramLine("Plaquetas 453.000")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Status != ParseStatusPartial {
		t.Fatalf("expected partial, got %s", got.Status)
	}
	if got.Code != "platelets" {
		t.Fatalf("expected platelets, got %q", got.Code)
	}
	if got.Value == nil || *got.Value != 453000 {
		t.Fatalf("expected value 453000, got %#v", got.Value)
	}
	if got.Unit != nil {
		t.Fatalf("expected nil unit, got %#v", got.Unit)
	}
}

func TestParseHemogramLine_ComplexReferenceKeepsRawReference(t *testing.T) {
	got, ok := ParseHemogramLine("- Neutrófilos 66,0 % (Referencia: 50 a 70 2000 a 7000)")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Code != "neutrophils" {
		t.Fatalf("expected neutrophils, got %q", got.Code)
	}
	if got.ReferenceMin != nil || got.ReferenceMax != nil {
		t.Fatalf("expected complex reference without min/max, got min=%#v max=%#v", got.ReferenceMin, got.ReferenceMax)
	}
	if got.RawReference == nil || *got.RawReference != "50 a 70 2000 a 7000" {
		t.Fatalf("expected raw reference preserved, got %#v", got.RawReference)
	}
}

func TestParseHemogramLine_UnknownAnalyteIsPartial(t *testing.T) {
	got, ok := ParseHemogramLine("- Campo Novo 123 mg/dL")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if got.Status != ParseStatusPartial {
		t.Fatalf("expected partial, got %s", got.Status)
	}
	if got.Code != "" {
		t.Fatalf("expected empty code, got %q", got.Code)
	}
}

func TestHemogramParser_Parse(t *testing.T) {
	parser := NewHemogramParser()
	output, err := parser.Parse(context.Background(), ParseInput{
		RawText: "HEMOGRAMA\n- Hemoglobina 15,1 g/dL (Referencia: 13,5 a 17,5 g/dL)\n- VPM 8,2 /fl",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ExamType != ExamTypeHemogram {
		t.Fatalf("expected hemogram, got %s", output.ExamType)
	}
	if len(output.Results) != 2 {
		t.Fatalf("expected 2 parsed results, got %d", len(output.Results))
	}
}

func TestHemogramParser_ParseNameValueBlocks(t *testing.T) {
	parser := NewHemogramParser()
	output, err := parser.Parse(context.Background(), ParseInput{
		RawText: `ERITROGRAMA
Hemácias
Hematócrito
Hemoglobina
VCM
HCM
CHCM
RDW

Valores de Referência:

4,98 milhões/ mm³
44,8 %
15,1 g/dL
90,0 fl
30,3 pg
33,7 g/dL
13,1 %`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedResult(t, output.Results, "rbc", 4.98, "milhões/ mm³")
	assertParsedResult(t, output.Results, "hematocrit", 44.8, "%")
	assertParsedResult(t, output.Results, "hemoglobin", 15.1, "g/dL")
	assertParsedResult(t, output.Results, "mcv", 90.0, "fl")
	assertParsedResult(t, output.Results, "mch", 30.3, "pg")
	assertParsedResult(t, output.Results, "mchc", 33.7, "g/dL")
	assertParsedResult(t, output.Results, "rdw", 13.1, "%")
}

func TestHemogramParser_ParseLeukogramNameValueBlocks(t *testing.T) {
	parser := NewHemogramParser()
	output, err := parser.Parse(context.Background(), ParseInput{
		RawText: `LEUCOGRAMA
Leucócitos
Neutrófilos
Promielocitos
Mielocitos
Metamielocitos
Bastões
Segmentados
Eosinofilos
Basofilos
Linfócitos típicos
Linfócitos atípicos
Monócitos
Blastos

8.600 /mm³
Percentual
66,0 %
0,0 %
0,0 %
0,0 %
0,0 %
66,0 %
4,0 %
0,0 %
24,0 %
0,0 %
6,0 %
0,0 %`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedResult(t, output.Results, "leukocytes", 8600, "/mm³")
	assertParsedResult(t, output.Results, "neutrophils", 66.0, "%")
	assertParsedResult(t, output.Results, "bands", 0.0, "%")
	assertParsedResult(t, output.Results, "lymphocytes", 24.0, "%")
	assertParsedResult(t, output.Results, "monocytes", 6.0, "%")
}

func assertParsedResult(t *testing.T, results []ParsedLabResult, code string, value float64, unit string) {
	t.Helper()

	for _, result := range results {
		if result.Code != code {
			continue
		}
		if result.Status != ParseStatusParsed {
			t.Fatalf("expected %s to be parsed, got %s", code, result.Status)
		}
		if result.Value == nil || *result.Value != value {
			t.Fatalf("expected %s value %v, got %#v", code, value, result.Value)
		}
		if result.Unit == nil || *result.Unit != unit {
			t.Fatalf("expected %s unit %q, got %#v", code, unit, result.Unit)
		}
		return
	}

	t.Fatalf("expected parsed result with code %q, got %#v", code, results)
}
