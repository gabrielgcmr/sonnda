package labparser

import "testing"

func TestParseReferenceRange_MinMax(t *testing.T) {
	got := ParseReferenceRange("13,5 a 17,5 g/dL")

	if got.Min == nil || *got.Min != 13.5 {
		t.Fatalf("expected min 13.5, got %#v", got.Min)
	}
	if got.Max == nil || *got.Max != 17.5 {
		t.Fatalf("expected max 17.5, got %#v", got.Max)
	}
	if got.Text != "13,5 a 17,5 g/dL" {
		t.Fatalf("expected raw text preserved, got %q", got.Text)
	}
}

func TestParseReferenceRange_ComplexKeepsTextOnly(t *testing.T) {
	got := ParseReferenceRange("50 a 70 2000 a 7000")

	if got.Min != nil || got.Max != nil {
		t.Fatalf("expected complex reference without min/max, got %#v", got)
	}
	if got.Text != "50 a 70 2000 a 7000" {
		t.Fatalf("expected raw text preserved, got %q", got.Text)
	}
}
