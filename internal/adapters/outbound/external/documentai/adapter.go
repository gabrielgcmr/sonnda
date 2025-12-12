package documentai

import (
	"context"
	"fmt"
	"sonnda-api/internal/core/ports/services"
)

// LabReportExtractor já está definido em labtest/ai.go
// type LabReportExtractor interface {
//     ExtractLabReport(ctx context.Context, gcsURI, mimeType string) (*ExtractedLabReport, error)
// }

// DocumentAIAdapter é a implementação de LabReportExtractor
// usando o client genérico de infra/docai.
type DocumentAIAdapter struct {
	client      Client
	processorID string
}

// Garante que implementa a interface
var _ services.DocumentExtractor = (*DocumentAIAdapter)(nil)

// NewDocumentAIAdapter é o construtor que você vai usar no module.go.
func NewDocumentAIAdapter(client Client, processorID string) *DocumentAIAdapter {
	return &DocumentAIAdapter{
		client:      client,
		processorID: processorID,
	}
}

func (a *DocumentAIAdapter) ExtractLabReport(
	ctx context.Context,
	documentURI, mimeType string,
) (*services.ExtractedLabReport, error) {
	// 1. Processa documento via Google Document AI
	doc, err := a.client.ProcessDocument(ctx, a.processorID, documentURI, mimeType)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar documento: %w", err)
	}

	// 2. Valida resposta
	if doc == nil {
		return nil, fmt.Errorf("documento retornado é nulo")
	}

	// 3. Converte Document protobuf → ExtractedLabReport
	extracted := mapDocumentToExtractedLabs(doc)

	// 4. Validação básica (opcional)
	if err := a.validateExtracted(extracted); err != nil {
		return nil, fmt.Errorf("validação falhou: %w", err)
	}

	return extracted, nil
}

func (a *DocumentAIAdapter) validateExtracted(extracted *services.ExtractedLabReport) error {
	// Você pode adicionar validações aqui se necessário
	// Por exemplo: garantir que pelo menos um teste foi extraído
	if len(extracted.Tests) == 0 {
		return fmt.Errorf("nenhum teste foi extraído do documento")
	}

	return nil
}

func (a *DocumentAIAdapter) ExtractImageExam(
	ctx context.Context,
	documentURI, mimeType string,
) (*services.ExtractedImageExam, error) {
	//TODO: Implementar
	return nil, nil

}
