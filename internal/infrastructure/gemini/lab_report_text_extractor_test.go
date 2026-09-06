// internal/infrastructure/gemini/lab_report_text_extractor_test.go
package gemini

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	"google.golang.org/genai"
)

func labReportExtractorTestClient(t *testing.T, generator ContentGenerator) *Client {
	t.Helper()
	cfg := testConfig()
	cfg.MaxInputBytes = 64 * 1024
	client, err := NewClientWithGenerator(cfg, generator)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func geminiTextResponse(text string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content:      genai.NewContentFromText(text, genai.RoleModel),
				FinishReason: genai.FinishReasonStop,
			},
		},
	}
}

func readExpectedLabJSON(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "domain", "labextraction", "testdata", filename))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestLabReportTextExtractorUsesPromptSchemaAndMapsResponse(t *testing.T) {
	expectedJSON := readExpectedLabJSON(t, "glicose.expected.json")
	var captured GenerateRequest
	client := labReportExtractorTestClient(t, generatorFunc(func(ctx context.Context, model string, contents []*genai.Content, cfg *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		if len(contents) != 1 {
			t.Fatal("unexpected content count")
		}
		captured = GenerateRequest{
			Text:              contents[0].Parts[0].Text,
			SystemInstruction: cfg.SystemInstruction.Parts[0].Text,
			JSONSchema:        cfg.ResponseJsonSchema.(map[string]any),
		}
		return geminiTextResponse(expectedJSON), nil
	}))
	extractor, err := NewLabReportTextExtractor(client)
	if err != nil {
		t.Fatal(err)
	}
	input := labextraction.ExtractLabReportInput{Text: "GLICEMIA DE JEJUM\nGlicemia de jejum 99 mg/dL"}
	report, err := extractor.ExtractLabReport(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if captured.Text != input.Text {
		t.Fatal("input text was not preserved")
	}
	if !strings.Contains(captured.SystemInstruction, "conteudo nao confiavel") {
		t.Fatal("prompt does not guard document instructions")
	}
	if _, exists := captured.JSONSchema["$schema"]; exists {
		t.Fatal("provider schema kept unsupported $schema")
	}
	if report.Metadata.Provider != "gemini" || report.Metadata.Model != "test-model" || report.Metadata.Status != labextraction.ExtractionStatusSucceeded {
		t.Fatalf("unexpected metadata: %+v", report.Metadata)
	}
	if report.RawText == nil || *report.RawText != input.Text {
		t.Fatal("raw input text was not attached")
	}
	if len(report.Tests) != 1 || len(report.Tests[0].Items) != 1 {
		t.Fatalf("unexpected tests: %+v", report.Tests)
	}
	item := report.Tests[0].Items[0]
	if item.ParameterName != "Glicemia de jejum" || item.ResultValue == nil || *item.ResultValue != "99" || item.ResultUnit == nil || *item.ResultUnit != "mg/dL" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if report.Tests[0].Status != labextraction.ExtractionStatusSucceeded || item.Status != labextraction.ExtractionStatusSucceeded {
		t.Fatal("nested status was not marked")
	}
}

func TestLabReportTextExtractorAcceptsNoStructuredResultsWithReviewStatus(t *testing.T) {
	client := labReportExtractorTestClient(t, generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		return geminiTextResponse(readExpectedLabJSON(t, "atestado.expected.json")), nil
	}))
	extractor, err := NewLabReportTextExtractor(client)
	if err != nil {
		t.Fatal(err)
	}
	report, err := extractor.ExtractLabReport(context.Background(), labextraction.ExtractLabReportInput{Text: "ATESTADO MEDICO"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Metadata.Status != labextraction.ExtractionStatusNeedsReview || len(report.Metadata.Warnings) != 1 {
		t.Fatalf("expected review status with warning, got %+v", report.Metadata)
	}
	if len(report.Tests) != 0 {
		t.Fatalf("unexpected tests: %+v", report.Tests)
	}
}

func TestLabReportTextExtractorValidatesJSONAndSchema(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		want    error
	}{
		{"invalid_json", "{", ErrInvalidResponseJSON},
		{"schema_error", `{"tests":[]}`, ErrResponseSchemaValidation},
		{"markdown_fence", "```json\n" + readExpectedLabJSON(t, "glicose.expected.json") + "\n```", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := labReportExtractorTestClient(t, generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
				return geminiTextResponse(tc.payload), nil
			}))
			extractor, err := NewLabReportTextExtractor(client)
			if err != nil {
				t.Fatal(err)
			}
			_, err = extractor.ExtractLabReport(context.Background(), labextraction.ExtractLabReportInput{Text: "Glicose 99 mg/dL"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestLabReportTextExtractorInterpretsProviderStopReasons(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason genai.FinishReason
		want   error
	}{
		{"max_tokens", genai.FinishReasonMaxTokens, ErrTruncatedResponse},
		{"safety", genai.FinishReasonSafety, ErrBlockedResponse},
		{"blocklist", genai.FinishReasonBlocklist, ErrBlockedResponse},
		{"other", genai.FinishReasonOther, ErrUnexpectedFinishReason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := geminiTextResponse(readExpectedLabJSON(t, "glicose.expected.json"))
			response.Candidates[0].FinishReason = tc.reason
			client := labReportExtractorTestClient(t, generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
				return response, nil
			}))
			extractor, err := NewLabReportTextExtractor(client)
			if err != nil {
				t.Fatal(err)
			}
			_, err = extractor.ExtractLabReport(context.Background(), labextraction.ExtractLabReportInput{Text: "Glicose 99 mg/dL"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestLabReportTextExtractorRejectsMissingCandidateOrText(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response *genai.GenerateContentResponse
		want     error
	}{
		{"nil_response", nil, ErrMissingResponseCandidate},
		{"no_candidate", &genai.GenerateContentResponse{}, ErrMissingResponseCandidate},
		{"empty_text", geminiTextResponse(" \n"), ErrEmptyResponseText},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := labReportExtractorTestClient(t, generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
				return tc.response, nil
			}))
			extractor, err := NewLabReportTextExtractor(client)
			if err != nil {
				t.Fatal(err)
			}
			_, err = extractor.ExtractLabReport(context.Background(), labextraction.ExtractLabReportInput{Text: "Glicose 99 mg/dL"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestLabReportTextExtractorPropagatesInputAndClientErrors(t *testing.T) {
	if _, err := NewLabReportTextExtractor(nil); !errors.Is(err, ErrMissingLabExtractionClient) {
		t.Fatalf("expected missing client error, got %v", err)
	}
	clientErr := errors.New("provider down")
	client := labReportExtractorTestClient(t, generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		return nil, clientErr
	}))
	extractor, err := NewLabReportTextExtractor(client)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := extractor.ExtractLabReport(context.Background(), labextraction.ExtractLabReportInput{}); !errors.Is(err, labextraction.ErrEmptyText) {
		t.Fatalf("expected input validation error, got %v", err)
	}
	if _, err := extractor.ExtractLabReport(context.Background(), labextraction.ExtractLabReportInput{Text: "Glicose 99 mg/dL"}); !errors.Is(err, clientErr) {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestProviderSchemaStripsUnsupportedValidationKeywords(t *testing.T) {
	providerSchema, _, err := loadLabReportSchemas()
	if err != nil {
		t.Fatal(err)
	}
	if containsSchemaKey(providerSchema, "$schema") || containsSchemaKey(providerSchema, "minLength") || containsSchemaKey(providerSchema, "pattern") {
		t.Fatal("provider schema kept unsupported local validation keywords")
	}
	if !reflect.DeepEqual(providerSchema["required"], []any{
		"patient_name", "patient_dob", "lab_name", "lab_phone",
		"insurance_provider", "requesting_doctor", "technical_manager", "report_date", "tests",
	}) {
		t.Fatal("provider schema lost required fields")
	}
}

func containsSchemaKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for k, child := range typed {
			if k == key || containsSchemaKey(child, key) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsSchemaKey(child, key) {
				return true
			}
		}
	}
	return false
}
