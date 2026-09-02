package labparser

import "testing"

func TestDetectExamType_Hemogram(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "title", text: "HEMOGRAMA\nMetodo: contagem automatizada"},
		{name: "erythrogram", text: "ERITROGRAMA\nHemacias\nHematocrito"},
		{name: "leukogram", text: "LEUCOGRAMA\nLeucocitos\nNeutrofilos"},
		{name: "platelets", text: "Plaquetas 453.000 /mm3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectExamType(tt.text)
			if got != ExamTypeHemogram {
				t.Fatalf("expected %s, got %s", ExamTypeHemogram, got)
			}
		})
	}
}

func TestDetectExamType_Unknown(t *testing.T) {
	got := DetectExamType("Glicose em jejum\nResultado 90 mg/dL")
	if got != ExamTypeUnknown {
		t.Fatalf("expected %s, got %s", ExamTypeUnknown, got)
	}
}
