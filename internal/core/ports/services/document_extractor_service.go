package services

import "context"

// ExtractedTestItem representa um test_item vindo do Document AI.
type ExtractedTestItem struct {
	ParameterName string  // test_item.parameter_name
	ResultText    *string // test_item.result
	ResultUnit    *string // test_item.unit
	ReferenceText *string // test_item.reference_text
}

// ExtractedTestResult representa um test_result vindo do Document AI.
type ExtractedTestResult struct {
	TestName    string              // (você provavelmente vai derivar do próprio painel)
	Material    *string             // test_result.material
	Method      *string             // test_result.method
	CollectedAt *string             // string de data/hora (vamos tratar depois)
	ReleaseAt   *string             // idem
	Items       []ExtractedTestItem // filhos
}

// ExtractedLabReport é o "DTO" vindo do Document AI já estruturado.
type ExtractedLabReport struct {
	PatientName       *string
	PatientDOB        *string // também como string por enquanto
	LabName           *string
	LabPhone          *string
	InsuranceProvider *string
	RequestingDoctor  *string
	TechnicalManager  *string
	ReportDate        *string
	RawText           *string

	Tests []ExtractedTestResult
}

type DocumentExtractor interface {
	ExtractLabReport(ctx context.Context, documentURI, mimeType string) (*ExtractedLabReport, error)
}
