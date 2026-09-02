package labparser

import "testing"

func TestNormalizeWhitespace(t *testing.T) {
	got := NormalizeWhitespace("  Hemoglobina\t15,1\r\n  g/dL  ")
	want := "Hemoglobina 15,1 g/dL"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNormalizeForMatch(t *testing.T) {
	got := NormalizeForMatch("Referência: milhões/ mm³\nLeucócitos")
	want := "referencia milhoes mm3 leucocitos"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseBrazilianDecimal(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "decimal comma", raw: "15,1", want: 15.1},
		{name: "thousands dot", raw: "453.000", want: 453000},
		{name: "thousands dot small", raw: "8.600", want: 8600},
		{name: "thousands and decimal", raw: "1.234,56", want: 1234.56},
		{name: "ocr l in decimal", raw: "15,l", want: 15.1},
		{name: "ocr o in thousands", raw: "453.OOO", want: 453000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBrazilianDecimal(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParseBrazilianDecimal_Invalid(t *testing.T) {
	if _, err := ParseBrazilianDecimal("hemoglobina"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeNumericToken_DoesNotRewriteTextOutsideNumericCall(t *testing.T) {
	got := NormalizeForMatch("COLESTEROL")
	want := "colesterol"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
