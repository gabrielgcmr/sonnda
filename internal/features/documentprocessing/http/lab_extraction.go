// internal/features/documentprocessing/http/lab_extraction.go
package http

import (
	"bytes"
	"context"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/extraction"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

const temporaryLabExtractionMaxFileSize = 10 * 1024 * 1024

type temporaryLabTextExtractor interface {
	Extract(ctx context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error)
}

type TemporaryLabExtractionHandler struct {
	extractor *extraction.Service
}

type temporaryLabExtractionForm struct {
	File huma.FormFile `form:"file" required:"true"`
}

type temporaryLabExtractionInput struct {
	RawBody huma.MultipartFormFiles[temporaryLabExtractionForm]
}

type temporaryLabExtractionResponse = extraction.Result

type temporaryLabExtractionOutput struct {
	Body temporaryLabExtractionResponse
}

func NewTemporaryLabExtraction(
	textExtractor temporaryLabTextExtractor,
	labExtractor labextraction.LabReportTextExtractor,
) *TemporaryLabExtractionHandler {
	return &TemporaryLabExtractionHandler{extractor: extraction.New(textExtractor, labExtractor)}
}

func (h *TemporaryLabExtractionHandler) RegisterHumaRoutes(registered huma.API, security []map[string][]string) {
	huma.Register(registered, huma.Operation{
		OperationID:  "extractTemporaryLabReport",
		Method:       http.MethodPost,
		Path:         "/lab-extractions",
		Summary:      "Extrair dados laboratoriais sem persistir o documento",
		MaxBodyBytes: 11 * 1024 * 1024,
		Tags:         []string{"Lab extraction"},
		Errors:       []int{http.StatusUnauthorized, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity},
		Security:     security,
	}, h.extract)
}

func (h *TemporaryLabExtractionHandler) extract(ctx context.Context, input *temporaryLabExtractionInput) (*temporaryLabExtractionOutput, error) {
	if _, err := humaCurrentUser(ctx); err != nil {
		return nil, err
	}

	files := input.RawBody.Form.File["file"]
	if len(files) != 1 {
		return nil, huma.Error422UnprocessableEntity("arquivo PDF e obrigatorio")
	}
	path, err := writeTemporaryPDF(files[0])
	if err != nil {
		return nil, toHumaError(ctx, err)
	}
	defer os.Remove(path)

	result, err := h.extractor.ExtractPDF(ctx, path, files[0].Filename)
	if err != nil {
		return nil, toHumaError(ctx, err)
	}
	return &temporaryLabExtractionOutput{Body: *result}, nil
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

	source, err := header.Open()
	if err != nil {
		return "", apperr.Internal("falha ao abrir arquivo", err)
	}
	defer source.Close()
	signature := make([]byte, 5)
	if _, err := io.ReadFull(source, signature); err != nil || !bytes.Equal(signature, []byte("%PDF-")) {
		return "", apperr.Validation("Envie um arquivo PDF válido.")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", apperr.Internal("falha ao ler arquivo", err)
	}
	target, err := os.CreateTemp("", "sonnda-lab-extraction-*.pdf")
	if err != nil {
		return "", apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	path := target.Name()
	count, err := io.Copy(target, io.LimitReader(source, temporaryLabExtractionMaxFileSize+1))
	if err != nil {
		target.Close()
		os.Remove(path)
		return "", apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	if count > temporaryLabExtractionMaxFileSize {
		_ = target.Close()
		_ = os.Remove(path)
		return "", &apperr.AppError{Kind: apperr.UPLOAD_SIZE_EXCEEDED, Message: "O PDF deve ter no máximo 10 MB."}
	}
	if err := target.Close(); err != nil {
		os.Remove(path)
		return "", apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	return path, nil
}
