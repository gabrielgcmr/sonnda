// internal/infrastructure/gemini/client.go
package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"google.golang.org/genai"
)

var (
	ErrMissingAPIKey = errors.New("GEMINI_API_KEY is required")
	ErrEmptyInput    = errors.New("gemini input text is required")
	ErrInputTooLarge = errors.New("gemini input exceeds configured byte limit")
)

// ContentGenerator permite simular o SDK sem chamadas externas.
type ContentGenerator interface {
	GenerateContent(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

var _ ContentGenerator = (*genai.Models)(nil)

type GenerateRequest struct {
	Text              string
	SystemInstruction string
	JSONSchema        map[string]any
}

type Client struct {
	generator ContentGenerator
	cfg       config.GeminiConfig
}

func NewClient(ctx context.Context, cfg config.GeminiConfig) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrMissingAPIKey
	}
	attempts := int32(1)
	sdk, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  strings.TrimSpace(cfg.APIKey),
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			Timeout:      &cfg.Timeout,
			RetryOptions: &genai.HTTPRetryOptions{Attempts: &attempts},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return NewClientWithGenerator(cfg, sdk.Models)
}

func NewClientWithGenerator(cfg config.GeminiConfig, generator ContentGenerator) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if generator == nil {
		return nil, errors.New("gemini content generator is required")
	}
	// A credencial fica somente no SDK, nao na configuracao retida pelo wrapper.
	cfg.APIKey = ""
	cfg.Model = strings.TrimSpace(cfg.Model)
	return &Client{generator: generator, cfg: cfg}, nil
}

// Generate devolve a resposta crua; interpretacao laboratorial fica no adaptador.
func (c *Client) Generate(ctx context.Context, request GenerateRequest) (*genai.GenerateContentResponse, error) {
	if strings.TrimSpace(request.Text) == "" {
		return nil, ErrEmptyInput
	}
	remaining := c.cfg.MaxInputBytes
	for _, part := range []string{request.Text, request.SystemInstruction} {
		if len(part) > remaining {
			return nil, ErrInputTooLarge
		}
		remaining -= len(part)
	}
	if request.JSONSchema != nil {
		schema, err := json.Marshal(request.JSONSchema)
		if err != nil {
			return nil, errors.New("gemini response schema is not valid JSON")
		}
		if len(schema) > remaining {
			return nil, ErrInputTooLarge
		}
	}
	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()
	if err := callCtx.Err(); err != nil {
		return nil, err
	}
	generation := &genai.GenerateContentConfig{
		CandidateCount:  1,
		MaxOutputTokens: c.cfg.MaxOutputTokens,
	}
	if request.SystemInstruction != "" {
		generation.SystemInstruction = genai.NewContentFromText(request.SystemInstruction, genai.RoleUser)
	}
	if request.JSONSchema != nil {
		generation.ResponseMIMEType = "application/json"
		generation.ResponseJsonSchema = request.JSONSchema
	}
	response, err := c.generator.GenerateContent(callCtx, c.cfg.Model, genai.Text(request.Text), generation)
	if err != nil {
		return nil, fmt.Errorf("gemini generate content: %w", err)
	}
	return response, nil
}
