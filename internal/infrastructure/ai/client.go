// internal/infrastructure/ai/client.go
package ai

import (
	"context"
	"fmt"
	"os"

	documentai "cloud.google.com/go/documentai/apiv1"
	"cloud.google.com/go/documentai/apiv1/documentaipb"
	"google.golang.org/api/option"
)

// Client é o contrato genérico de acesso ao Document AI.
// Ele não conhece nenhum domínio (labs, examImage, etc.):
// recebe um processorID e uma URI do GCS e devolve o Document cru.
type Client struct {
	client    *documentai.DocumentProcessorClient
	projectID string
	location  string
}

func NewClient(
	ctx context.Context,
	projectID, location string,
	opts ...option.ClientOption,
) (*Client, error) {
	// Se opts estiver vazio, tenta configurar credenciais automaticamente
	if len(opts) == 0 {
		if credsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"); credsJSON != "" {
			// Usa JSON inline (production/Railway)
			opts = append(opts, option.WithCredentialsJSON([]byte(credsJSON)))
		} else if credsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credsFile != "" {
			// Usa arquivo local (desenvolvimento)
			opts = append(opts, option.WithCredentialsFile(credsFile))
		}
	}

	c, err := documentai.NewDocumentProcessorClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente do Document AI: %w", err)
	}

	return &Client{
		client:    c,
		projectID: projectID,
		location:  location,
	}, nil
}

func (c *Client) ProcessDocument(
	ctx context.Context,
	processorID, gcsURI, mimeType string,
) (*documentaipb.Document, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/processors/%s", c.projectID, c.location, processorID)

	req := &documentaipb.ProcessRequest{
		Name: name,
		Source: &documentaipb.ProcessRequest_GcsDocument{
			GcsDocument: &documentaipb.GcsDocument{
				GcsUri:   gcsURI,
				MimeType: mimeType,
			},
		},
	}

	resp, err := c.client.ProcessDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("falha no processamento do DocAI: %w", err)
	}

	if resp.Document == nil {
		return nil, fmt.Errorf("document AI retornou resposta sem documento")
	}

	return resp.Document, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
