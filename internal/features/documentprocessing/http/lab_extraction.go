// internal/features/documentprocessing/http/lab_extraction.go
package http

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/api/humaerror"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

const temporaryLabExtractionMaxFileSize = 10 * 1024 * 1024

type temporaryLabTextExtractor interface {
	Extract(ctx context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error)
}

type TemporaryLabExtractionHandler struct {
	textExtractor temporaryLabTextExtractor
	labExtractor  labextraction.LabReportTextExtractor
}

type temporaryLabExtractionForm struct {
	File huma.FormFile `form:"file" required:"true"`
}

type temporaryLabExtractionInput struct {
	RawBody huma.MultipartFormFiles[temporaryLabExtractionForm]
}

type temporaryLabExtractionResponse struct {
	Status   labextraction.ExtractionStatus    `json:"status"`
	Warnings []labextraction.ExtractionWarning `json:"warnings,omitempty"`
	Report   labextraction.ExtractedLabReport  `json:"report"`
}

type temporaryLabExtractionOutput struct {
	Body temporaryLabExtractionResponse
}

func NewTemporaryLabExtraction(
	textExtractor temporaryLabTextExtractor,
	labExtractor labextraction.LabReportTextExtractor,
) *TemporaryLabExtractionHandler {
	return &TemporaryLabExtractionHandler{textExtractor: textExtractor, labExtractor: labExtractor}
}

func (h *TemporaryLabExtractionHandler) RegisterHumaRoutes(registered huma.API, security []map[string][]string) {
	huma.Register(registered, huma.Operation{
		OperationID: "extractTemporaryLabReport",
		Method:      http.MethodPost,
		Path:        "/lab-extractions",
		Summary:     "Extrair dados laboratoriais sem persistir o documento",
		Tags:        []string{"Lab extraction"},
		Errors:      []int{http.StatusUnauthorized, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity},
		Security:    security,
	}, h.extract)
}

func (h *TemporaryLabExtractionHandler) extract(ctx context.Context, input *temporaryLabExtractionInput) (*temporaryLabExtractionOutput, error) {
	if _, err := humaCurrentUser(ctx); err != nil {
		return nil, err
	}
	if isNilExtractor(h.textExtractor) || isNilExtractor(h.labExtractor) {
		return nil, humaerror.From(apperr.Internal("A extracao laboratorial esta indisponivel.", nil))
	}

	files := input.RawBody.Form.File["file"]
	if len(files) != 1 {
		return nil, huma.Error422UnprocessableEntity("arquivo PDF e obrigatorio")
	}
	path, err := writeTemporaryPDF(files[0])
	if err != nil {
		return nil, humaerror.From(err)
	}
	defer os.Remove(path)

	text, err := h.textExtractor.Extract(ctx, domaintext.ExtractInput{
		LocalPath: path, MimeType: "application/pdf", OriginalFilename: files[0].Filename,
	})
	if err != nil || text == nil || strings.TrimSpace(text.Text) == "" {
		return nil, humaerror.From(apperr.DomainRuleViolation("Nao foi encontrado texto legivel no PDF. Envie o PDF original com texto selecionavel.", apperr.Violation{Field: "file", Reason: "text_not_readable"}))
	}

	report, err := h.labExtractor.ExtractLabReport(ctx, labextraction.ExtractLabReportInput{Text: text.Text})
	if err != nil {
		return nil, humaerror.From(apperr.Internal("Nao foi possivel extrair os dados laboratoriais.", err))
	}
	if report == nil {
		return nil, humaerror.From(apperr.Internal("Nao foi possivel extrair os dados laboratoriais.", nil))
	}
	report.Normalize()
	status := report.Metadata.Status
	if status == "" {
		status = labextraction.ExtractionStatusSucceeded
	}
	if !report.HasStructuredResults() && status == labextraction.ExtractionStatusSucceeded {
		status = labextraction.ExtractionStatusNeedsReview
	}
	return &temporaryLabExtractionOutput{Body: temporaryLabExtractionResponse{
		Status: status, Warnings: report.Metadata.Warnings, Report: *report,
	}}, nil
}

func isNilExtractor(extractor any) bool {
	if extractor == nil {
		return true
	}
	value := reflect.ValueOf(extractor)
	return value.Kind() == reflect.Ptr && value.IsNil()
}

func writeTemporaryPDF(header *multipart.FileHeader) (string, error) {
	if header == nil {
		return "", apperr.Validation("arquivo PDF e obrigatorio", apperr.Violation{Field: "file", Reason: "required"})
	}
	if header.Size <= 0 {
		return "", apperr.Validation("arquivo vazio", apperr.Violation{Field: "file", Reason: "empty"})
	}
	if header.Size > temporaryLabExtractionMaxFileSize {
		return "", &apperr.AppError{Kind: apperr.UPLOAD_SIZE_EXCEEDED, Message: "o PDF deve ter no maximo 10 MB"}
	}
	if normalizeMimeType(header.Header.Get("Content-Type")) != "application/pdf" && strings.ToLower(filepath.Ext(header.Filename)) != ".pdf" {
		return "", &apperr.AppError{Kind: apperr.INVALID_FIELD_FORMAT, Message: "envie um arquivo PDF"}
	}

	source, err := header.Open()
	if err != nil {
		return "", apperr.Internal("falha ao abrir arquivo", err)
	}
	defer source.Close()
	target, err := os.CreateTemp("", "sonnda-lab-extraction-*.pdf")
	if err != nil {
		return "", apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	path := target.Name()
	if _, err := io.Copy(target, source); err != nil {
		target.Close()
		os.Remove(path)
		return "", apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	if err := target.Close(); err != nil {
		os.Remove(path)
		return "", apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	return path, nil
}
