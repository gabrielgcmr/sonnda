// internal/infrastructure/documentai/mapper_test.go
package documentai

import (
	"testing"

	"cloud.google.com/go/documentai/apiv1/documentaipb"
)

func TestMapDocumentToExtractedLabsSupportsPanelObservationSchema(t *testing.T) {
	doc := &documentaipb.Document{
		Text: "Hemograma Hemacias 4.8 milhoes/mm3",
		Entities: []*documentaipb.Document_Entity{
			{
				Type: "panel",
				Properties: []*documentaipb.Document_Entity{
					{Type: "test_name", MentionText: "Hemograma"},
					{Type: "material", MentionText: "Sangue"},
					{Type: "collected_at", NormalizedValue: &documentaipb.Document_Entity_NormalizedValue{Text: "2026-09-25"}},
					{
						Type: "observation",
						Properties: []*documentaipb.Document_Entity{
							{Type: "parameter_name", MentionText: "Hemacias"},
							{Type: "value_number", MentionText: "4.8"},
							{Type: "unit", MentionText: "milhoes/mm3"},
							{Type: "reference_text", MentionText: "4.0 a 5.4"},
						},
					},
				},
			},
		},
	}

	report := mapDocumentToExtractedLabs(doc)
	if len(report.Tests) != 1 {
		t.Fatalf("tests = %d, want 1", len(report.Tests))
	}
	test := report.Tests[0]
	if test.TestName != "Hemograma" || test.Material == nil || *test.Material != "Sangue" {
		t.Fatalf("mapped test = %+v", test)
	}
	if test.CollectedAt == nil || *test.CollectedAt != "2026-09-25" {
		t.Fatalf("collected_at = %v, want normalized date", test.CollectedAt)
	}
	if len(test.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(test.Items))
	}
	item := test.Items[0]
	if item.ParameterName != "Hemacias" || item.ResultValue == nil || *item.ResultValue != "4.8" {
		t.Fatalf("mapped item = %+v", item)
	}
	if !report.HasStructuredResults() {
		t.Fatal("new schema should produce structured results")
	}
}

func TestMapTestItemPreservesNumericValueWhenTextValueAlsoExists(t *testing.T) {
	item := mapTestItem(nil, &documentaipb.Document_Entity{
		Properties: []*documentaipb.Document_Entity{
			{Type: "value_number", MentionText: "90"},
			{Type: "value_text", MentionText: "normal"},
		},
	})

	if item.ResultValue == nil || *item.ResultValue != "90" {
		t.Fatalf("result value = %v, want numeric value", item.ResultValue)
	}
}
