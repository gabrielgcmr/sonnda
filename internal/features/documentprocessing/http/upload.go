// internal/features/documentprocessing/http/upload.go
package http

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	stdhttp "net/http"
	"os"
	"path/filepath"
	"strings"

	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type UploadedDocument struct {
	StorageURI       string
	OriginalFilename string
	MimeType         string
	LocalPath        string
}

func UploadDocument(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	patientID uuid.UUID,
	storage domainstorage.FileStorageService,
) (*UploadedDocument, error) {
	const maxFileSize = 10 * 1024 * 1024
	if fileHeader == nil {
		return nil, &apperr.AppError{Kind: apperr.REQUIRED_FIELD_MISSING, Message: "arquivo e obrigatorio"}
	}
	if fileHeader.Size == 0 {
		return nil, &apperr.AppError{Kind: apperr.VALIDATION_FAILED, Message: "arquivo vazio"}
	}
	if fileHeader.Size > maxFileSize {
		return nil, &apperr.AppError{Kind: apperr.UPLOAD_SIZE_EXCEEDED, Message: "arquivo muito grande"}
	}
	if patientID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}
	if storage == nil {
		return nil, apperr.Internal("armazenamento de documentos indisponivel", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apperr.Internal("falha ao abrir arquivo", err)
	}
	defer file.Close()
	tempFile, err := os.CreateTemp("", "sonnda-exam-*"+filepath.Ext(fileHeader.Filename))
	if err != nil {
		return nil, apperr.Internal("falha ao preparar arquivo temporario", err)
	}
	tempPath := tempFile.Name()
	removeTempOnError := true
	defer func() {
		if removeTempOnError {
			_ = os.Remove(tempPath)
		}
	}()
	if _, err := io.Copy(tempFile, file); err != nil {
		_ = tempFile.Close()
		return nil, apperr.Internal("falha ao copiar arquivo", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, apperr.Internal("falha ao fechar arquivo temporario", err)
	}

	contentType, err := contentTypeForUpload(fileHeader, tempPath)
	if err != nil {
		return nil, err
	}
	ext := mimeToExtension(contentType)
	objectName := fmt.Sprintf("patients/%s/exam-documents/%s%s", patientID, uuid.NewString(), ext)
	uploadFile, err := os.Open(tempPath)
	if err != nil {
		return nil, apperr.Internal("falha ao reabrir arquivo", err)
	}
	defer uploadFile.Close()
	uri, err := storage.Upload(ctx, uploadFile, objectName, contentType)
	if err != nil {
		return nil, &apperr.AppError{Kind: apperr.INFRA_STORAGE_ERROR, Message: "falha no upload", Cause: err}
	}
	removeTempOnError = false
	return &UploadedDocument{StorageURI: uri, OriginalFilename: fileHeader.Filename, MimeType: contentType, LocalPath: tempPath}, nil
}

func contentTypeForUpload(fileHeader *multipart.FileHeader, tempPath string) (string, error) {
	contentType := normalizeMimeType(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		detectFile, err := os.Open(tempPath)
		if err != nil {
			return "", apperr.Internal("falha ao detectar tipo de arquivo", err)
		}
		defer detectFile.Close()
		buffer := make([]byte, 512)
		count, _ := detectFile.Read(buffer)
		contentType = normalizeMimeType(stdhttp.DetectContentType(buffer[:count]))
	}
	if !isSupportedMimeType(contentType) {
		return "", &apperr.AppError{Kind: apperr.INVALID_FIELD_FORMAT, Message: "tipo de arquivo nao suportado", Cause: fmt.Errorf("content_type=%s", contentType)}
	}
	return contentType, nil
}

func isSupportedMimeType(contentType string) bool {
	switch contentType {
	case "application/pdf", "image/jpeg", "image/png":
		return true
	default:
		return false
	}
}

func normalizeMimeType(raw string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(raw, ";")[0]))
}

func mimeToExtension(contentType string) string {
	switch contentType {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ""
	}
}
