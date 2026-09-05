package labextraction

import "strings"

type ExtractionStatus string

const (
	ExtractionStatusSucceeded   ExtractionStatus = "succeeded"
	ExtractionStatusPartial     ExtractionStatus = "partial"
	ExtractionStatusNeedsReview ExtractionStatus = "needs_review"
	ExtractionStatusFailed      ExtractionStatus = "failed"
)

type ExtractionWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type ExtractionMetadata struct {
	Provider   string              `json:"provider,omitempty"`
	Processor  string              `json:"processor,omitempty"`
	Model      string              `json:"model,omitempty"`
	Status     ExtractionStatus    `json:"status,omitempty"`
	Confidence *float64            `json:"confidence,omitempty"`
	Warnings   []ExtractionWarning `json:"warnings,omitempty"`
}

// ExtractedTestItem representa um item estruturado vindo do extrator semantico.
type ExtractedTestItem struct {
	ParameterName string
	ResultValue   *string
	ResultUnit    *string
	ReferenceText *string
	RawText       *string
	Status        ExtractionStatus
	Confidence    *float64
	Warnings      []ExtractionWarning
}

// ExtractedTestResult representa um exame ou painel estruturado.
type ExtractedTestResult struct {
	TestName    string
	Material    *string
	Method      *string
	CollectedAt *string
	ReleaseAt   *string
	RawText     *string
	Status      ExtractionStatus
	Confidence  *float64
	Warnings    []ExtractionWarning
	Items       []ExtractedTestItem
}

// ExtractedLabReport e o contrato interno da extracao laboratorial.
type ExtractedLabReport struct {
	Metadata ExtractionMetadata

	PatientName       *string
	PatientDOB        *string
	LabName           *string
	LabPhone          *string
	InsuranceProvider *string
	RequestingDoctor  *string
	TechnicalManager  *string
	ReportDate        *string
	RawText           *string

	Tests []ExtractedTestResult
}

func (r *ExtractedLabReport) Normalize() {
	if r == nil {
		return
	}

	r.Metadata.Provider = strings.TrimSpace(r.Metadata.Provider)
	r.Metadata.Processor = strings.TrimSpace(r.Metadata.Processor)
	r.Metadata.Model = strings.TrimSpace(r.Metadata.Model)
	r.PatientName = trimStringPtr(r.PatientName)
	r.PatientDOB = trimStringPtr(r.PatientDOB)
	r.LabName = trimStringPtr(r.LabName)
	r.LabPhone = trimStringPtr(r.LabPhone)
	r.InsuranceProvider = trimStringPtr(r.InsuranceProvider)
	r.RequestingDoctor = trimStringPtr(r.RequestingDoctor)
	r.TechnicalManager = trimStringPtr(r.TechnicalManager)
	r.ReportDate = trimStringPtr(r.ReportDate)
	r.RawText = trimStringPtr(r.RawText)

	for testIndex := range r.Tests {
		r.Tests[testIndex].Normalize()
	}
}

func (r *ExtractedLabReport) HasStructuredResults() bool {
	if r == nil {
		return false
	}

	for _, test := range r.Tests {
		if strings.TrimSpace(test.TestName) == "" {
			continue
		}
		for _, item := range test.Items {
			if strings.TrimSpace(item.ParameterName) != "" {
				return true
			}
		}
	}

	return false
}

func (t *ExtractedTestResult) Normalize() {
	if t == nil {
		return
	}

	t.TestName = strings.TrimSpace(t.TestName)
	t.Material = trimStringPtr(t.Material)
	t.Method = trimStringPtr(t.Method)
	t.CollectedAt = trimStringPtr(t.CollectedAt)
	t.ReleaseAt = trimStringPtr(t.ReleaseAt)
	t.RawText = trimStringPtr(t.RawText)

	for itemIndex := range t.Items {
		t.Items[itemIndex].Normalize()
	}
}

func (i *ExtractedTestItem) Normalize() {
	if i == nil {
		return
	}

	i.ParameterName = strings.TrimSpace(i.ParameterName)
	i.ResultValue = trimStringPtr(i.ResultValue)
	i.ResultUnit = trimStringPtr(i.ResultUnit)
	i.ReferenceText = trimStringPtr(i.ReferenceText)
	i.RawText = trimStringPtr(i.RawText)
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
