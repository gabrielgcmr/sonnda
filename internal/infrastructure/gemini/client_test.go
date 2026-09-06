// internal/infrastructure/gemini/client_test.go
package gemini

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"google.golang.org/genai"
)

type generatorFunc func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)

func (f generatorFunc) GenerateContent(ctx context.Context, model string, contents []*genai.Content, cfg *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return f(ctx, model, contents, cfg)
}

func testConfig() config.GeminiConfig {
	return config.GeminiConfig{Model: "test-model", Timeout: time.Second, MaxInputBytes: 1024, MaxOutputTokens: 256}
}

func TestClientConfigurationAndRequest(t *testing.T) {
	request := GenerateRequest{
		Text: " Glicose 99 mg/dL\n", SystemInstruction: "Retorne JSON.",
		JSONSchema: map[string]any{"type": "object"},
	}
	response := &genai.GenerateContentResponse{}
	called := false
	client, err := NewClientWithGenerator(testConfig(), generatorFunc(func(ctx context.Context, model string, contents []*genai.Content, cfg *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		called = true
		if model != "test-model" || cfg.MaxOutputTokens != 256 || cfg.CandidateCount != 1 {
			t.Fatal("model or generation limits were not passed to SDK")
		}
		if len(contents) != 1 || contents[0].Parts[0].Text != request.Text {
			t.Fatal("input text was changed or truncated")
		}
		if cfg.SystemInstruction.Parts[0].Text != request.SystemInstruction || cfg.ResponseMIMEType != "application/json" || !reflect.DeepEqual(cfg.ResponseJsonSchema, request.JSONSchema) {
			t.Fatal("instructions or schema were not passed to SDK")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("generation has no deadline")
		}
		return response, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.Generate(context.Background(), request)
	if err != nil || got != response || !called {
		t.Fatalf("unexpected generation result: %v", err)
	}
}

func TestClientRejectsInputBeforeCallingSDK(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request GenerateRequest
		limit   int
		want    error
	}{
		{"empty", GenerateRequest{Text: " \n"}, 100, ErrEmptyInput},
		{"text_limit", GenerateRequest{Text: "12345"}, 4, ErrInputTooLarge},
		{"utf8_bytes", GenerateRequest{Text: "\u00e1"}, 1, ErrInputTooLarge},
		{"instruction_limit", GenerateRequest{Text: "123", SystemInstruction: "45"}, 4, ErrInputTooLarge},
		{"schema_limit", GenerateRequest{Text: "1", JSONSchema: map[string]any{"type": "object"}}, 4, ErrInputTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.MaxInputBytes = tc.limit
			client, err := NewClientWithGenerator(cfg, generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
				t.Fatal("SDK called for invalid input")
				return nil, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Generate(context.Background(), tc.request); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestClientTimeoutAndParentDeadline(t *testing.T) {
	for _, shorterParent := range []bool{false, true} {
		cfg := testConfig()
		cfg.Timeout = 20 * time.Millisecond
		parent := context.Background()
		cancel := func() {}
		if shorterParent {
			cfg.Timeout = time.Second
			parent, cancel = context.WithTimeout(parent, 20*time.Millisecond)
		}
		client, err := NewClientWithGenerator(cfg, generatorFunc(func(ctx context.Context, _ string, _ []*genai.Content, _ *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}))
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		_, err = client.Generate(parent, GenerateRequest{Text: "test"})
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline error, got %v", err)
		}
	}
}

func TestClientCancelledContextDoesNotCallSDK(t *testing.T) {
	client, err := NewClientWithGenerator(testConfig(), generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		t.Fatal("SDK called with cancelled context")
		return nil, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Generate(ctx, GenerateRequest{Text: "test"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestClientPreservesSDKErrorWithoutRetry(t *testing.T) {
	want := errors.New("simulated provider failure")
	calls := 0
	client, err := NewClientWithGenerator(testConfig(), generatorFunc(func(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		calls++
		return nil, want
	}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Generate(context.Background(), GenerateRequest{Text: "test"}); !errors.Is(err, want) || calls != 1 {
		t.Fatalf("unexpected error or retry count: %v, calls=%d", err, calls)
	}
}

func TestNewClientRequiresExplicitKey(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "unrelated-test-key")
	t.Setenv("GEMINI_API_KEY", "unrelated-test-key")
	if _, err := NewClient(context.Background(), testConfig()); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected explicit key requirement, got %v", err)
	}
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	cfg := testConfig()
	cfg.APIKey = "fake-key-not-sent"
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if client.cfg.APIKey != "" {
		t.Fatal("wrapper retained the credential")
	}
}

func TestNewClientRejectsInvalidConfiguration(t *testing.T) {
	if _, err := NewClientWithGenerator(testConfig(), nil); err == nil {
		t.Fatal("nil generator accepted")
	}
	cfg := testConfig()
	cfg.Timeout = 0
	if _, err := NewClient(context.Background(), cfg); err == nil {
		t.Fatal("invalid timeout accepted")
	}
}
