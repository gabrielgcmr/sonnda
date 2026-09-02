package labparser

import "testing"

func TestResolveAnalyteCode_HemogramAliases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "hemoglobin", raw: "Hemoglobina", want: "hemoglobin"},
		{name: "hgb", raw: "HGB", want: "hemoglobin"},
		{name: "hematocrit accented", raw: "Hematócrito", want: "hematocrit"},
		{name: "rbc accented", raw: "Hemácias", want: "rbc"},
		{name: "leukocytes accented", raw: "Leucócitos", want: "leukocytes"},
		{name: "platelets", raw: "Plaquetas", want: "platelets"},
		{name: "typical lymphocytes", raw: "Linfócitos típicos", want: "lymphocytes"},
		{name: "atypical lymphocytes", raw: "Linfócitos atípicos", want: "atypical_lymphocytes"},
		{name: "bands", raw: "Bastões", want: "bands"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ResolveAnalyteCode(tt.raw)
			if !ok {
				t.Fatalf("expected alias %q to resolve", tt.raw)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestResolveAnalyteCode_Unknown(t *testing.T) {
	if got, ok := ResolveAnalyteCode("Solicitante"); ok {
		t.Fatalf("expected unknown alias, got %q", got)
	}
}
