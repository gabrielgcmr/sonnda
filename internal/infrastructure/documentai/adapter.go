// internal/infrastructure/documentai/adapter.go
package documentai

import (
	"context"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
)

// DocumentAIAdapter adapta o Google Document AI ao contrato de laudos laboratoriais.
type DocumentAIAdapter struct {
	client      Client
	processorID string
}

var _ labextraction.LabReportExtractor = (*DocumentAIAdapter)(nil)

func NewDocumentAIAdapter(client Client, processorID string) *DocumentAIAdapter {
	return &DocumentAIAdapter{
		client:      client,
		processorID: processorID,
	}
}

func (a *DocumentAIAdapter) ExtractLabReport(
	ctx context.Context,
	documentURI, mimeType string,
) (*labextraction.ExtractedLabReport, error) {
	doc, err := a.client.ProcessDocument(ctx, a.processorID, documentURI, mimeType)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar documento: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("documento retornado é nulo")
	}

	extracted := mapDocumentToExtractedLabs(doc)
	extracted.Metadata.Provider = "document_ai"
	extracted.Metadata.Processor = a.processorID
	extracted.Metadata.Status = labextraction.ExtractionStatusSucceeded
	extracted.Normalize()

	if err := a.validateExtracted(extracted); err != nil {
		return nil, fmt.Errorf("validação falhou: %w", err)
	}

	return extracted, nil
}

func (a *DocumentAIAdapter) validateExtracted(extracted *labextraction.ExtractedLabReport) error {
	if extracted == nil || !extracted.HasStructuredResults() {
		return fmt.Errorf("nenhum teste foi extraído do documento")
	}

	return nil
}
