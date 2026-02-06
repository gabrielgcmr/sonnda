// internal/api/handlers/labs.go
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type LabsHandler struct {
	svc      labsvc.Service
	createUC labsuc.CreateLabReportFromDocumentUseCase
	storage  domainstorage.FileStorageService
}

func NewLabs(
	svc labsvc.Service,
	createUC labsuc.CreateLabReportFromDocumentUseCase,
	storageClient domainstorage.FileStorageService,
) *LabsHandler {
	return &LabsHandler{
		svc:      svc,
		createUC: createUC,
		storage:  storageClient,
	}
}

func (h *LabsHandler) ListLabs(c *gin.Context) {
	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}

	if shouldReturnFullLabs(c) {
		list, err := h.svc.ListFull(c.Request.Context(), patientID, limit, offset)
		if err != nil {
			presenter.ErrorResponder(c, err)
			return
		}

		c.JSON(http.StatusOK, list)
		return
	}

	list, err := h.svc.List(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

// Handler unico para upload de laudo
// POST /:patientID/labs
// field: file (PDF/JPEG/PNG)
func (h *LabsHandler) UploadAndProcessLabs(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	documentURI, mimeType, uploadErr := h.handleFileUpload(c, patientID)
	if uploadErr != nil {
		presenter.ErrorResponder(c, uploadErr)
		return
	}

	output, uploadErr := h.createUC.Execute(c.Request.Context(), labsuc.CreateLabReportFromDocumentInput{
		PatientID:        patientID,
		DocumentURI:      documentURI,
		MimeType:         mimeType,
		UploadedByUserID: currentUser.ID,
	})
	if uploadErr != nil {
		presenter.ErrorResponder(c, uploadErr)
		return
	}

	c.JSON(http.StatusCreated, output)
}

// handleFileUpload centraliza toda a logica de:
// - ler o arquivo do multipart
// - detectar/validar content-type
// - fazer upload pro storage
// - retornar (URI, MIME)
func (h *LabsHandler) handleFileUpload(
	c *gin.Context,
	patientID uuid.UUID,
) (string, string, error) {
	const MaxFileSize = 10 * 1024 * 1024 // 10MB

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return "", "", &apperr.AppError{
			Kind:    apperr.REQUIRED_FIELD_MISSING,
			Message: "arquivo é obrigatório",
			Cause:   err,
		}
	}
	if fileHeader.Size == 0 {
		return "", "", &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "arquivo vazio",
		}
	}

	if fileHeader.Size > MaxFileSize {
		return "", "", &apperr.AppError{
			Kind:    apperr.UPLOAD_SIZE_EXCEEDED,
			Message: "arquivo muito grande",
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", apperr.Internal("falha ao abrir arquivo", err)
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		contentType = http.DetectContentType(buf[:n])

		if seeker, ok := file.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
	}

	contentType = normalizeMimeType(contentType)

	if !isSupportedMimeType(contentType) {
		return "", "", &apperr.AppError{
			Kind:    apperr.INVALID_FIELD_FORMAT,
			Message: "tipo de arquivo não suportado",
			Cause:   fmt.Errorf("content_type=%s", contentType),
		}
	}

	uniqueID := uuid.NewString()
	ext := mimeToExt(contentType)
	if ext == "" {
		return "", "", &apperr.AppError{
			Kind:    apperr.INVALID_FIELD_FORMAT,
			Message: "tipo de arquivo não suportado",
			Cause:   fmt.Errorf("content_type=%s", contentType),
		}
	}

	if patientID == uuid.Nil {
		return "", "", apperr.Validation("entrada inválida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}

	objectName := fmt.Sprintf("patients/%s/lab-reports/%s%s", patientID.String(), uniqueID, ext)

	uri, err := h.storage.Upload(c.Request.Context(), file, objectName, contentType)
	if err != nil {
		return "", "", &apperr.AppError{
			Kind:    apperr.INFRA_STORAGE_ERROR,
			Message: "falha no upload",
			Cause:   err,
		}
	}

	return uri, contentType, nil
}

func parsePagination(c *gin.Context, defaultLimit, defaultOffset int) (limit, offset int, ok bool) {
	limit = defaultLimit
	offset = defaultOffset

	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			presenter.ErrorResponder(c, &apperr.AppError{
				Kind:    apperr.VALIDATION_FAILED,
				Message: "limit deve ser > 0",
				Cause:   err,
			})
			return 0, 0, false
		}
		limit = l
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			presenter.ErrorResponder(c, &apperr.AppError{
				Kind:    apperr.VALIDATION_FAILED,
				Message: "offset deve ser >= 0",
				Cause:   err,
			})
			return 0, 0, false
		}
		offset = o
	}

	return limit, offset, true
}

func shouldReturnFullLabs(c *gin.Context) bool {
	if strings.EqualFold(strings.TrimSpace(c.Query("expand")), "full") {
		return true
	}

	include := strings.TrimSpace(c.Query("include"))
	if include == "" {
		return false
	}

	for _, raw := range strings.Split(include, ",") {
		value := strings.ToLower(strings.TrimSpace(raw))
		switch value {
		case "full", "results", "test_results":
			return true
		}
	}

	return false
}

// isSupportedMimeType checks whether the upload is of an accepted type.
func isSupportedMimeType(ct string) bool {
	ct = strings.ToLower(ct)
	switch ct {
	case "application/pdf", "image/pdf":
		return true
	case "image/jpeg", "image/jpg":
		return true
	case "image/png":
		return true
	default:
		return false
	}
}

func normalizeMimeType(raw string) string {
	if raw == "" {
		return ""
	}
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.Index(normalized, ";"); idx >= 0 {
		normalized = strings.TrimSpace(normalized[:idx])
	}
	return normalized
}

func mimeToExt(ct string) string {
	switch strings.ToLower(strings.TrimSpace(ct)) {
	case "application/pdf", "image/pdf":
		return ".pdf"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ""
	}
}
