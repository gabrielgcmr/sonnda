// internal/features/documentprocessing/extraction/results.go
package extraction

import "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"

// Remove unnamed entries before preview, so the persisted report matches what was reviewed.
func normalizeResults(report *labextraction.ExtractedLabReport) {
	tests := make([]labextraction.ExtractedTestResult, 0, len(report.Tests))
	for _, test := range report.Tests {
		if test.TestName == "" {
			report.Metadata.Warnings = append(report.Metadata.Warnings, labextraction.ExtractionWarning{Code: "unnamed_test", Message: "Um exame sem nome identificado foi omitido.", Field: "tests"})
			continue
		}
		report.Metadata.Warnings = append(report.Metadata.Warnings, test.Warnings...)
		items := make([]labextraction.ExtractedTestItem, 0, len(test.Items))
		for _, item := range test.Items {
			if item.ParameterName == "" {
				report.Metadata.Warnings = append(report.Metadata.Warnings, labextraction.ExtractionWarning{Code: "unnamed_parameter", Message: "Um parâmetro sem nome identificado foi omitido.", Field: "tests.items"})
				continue
			}
			report.Metadata.Warnings = append(report.Metadata.Warnings, item.Warnings...)
			items = append(items, item)
		}
		if len(items) == 0 {
			continue
		}
		test.Items = items
		tests = append(tests, test)
	}
	report.Tests = tests
}
