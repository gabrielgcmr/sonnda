package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	labsvc "sonnda-api/internal/app/services/labs"
	labsuc "sonnda-api/internal/app/usecase/labs"

	"sonnda-api/internal/adapters/inbound/http/httperr"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/app/apperr"
	external "sonnda-api/internal/domain/ports/integration"
)

type LabsHandler struct {
	svc      labsvc.Service
	createUC labsuc.CreateLabReportFromDocumentUseCase
	storage  external.StorageService
}

func NewLabsHandler(
	svc labsvc.Service,
	createUC labsuc.CreateLabReportFromDocumentUseCase,
	storageClient external.StorageService,
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

	list, err := h.svc.List(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		httperr.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *LabsHandler) ListFullLabs(c *gin.Context) {
	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}

	list, err := h.svc.ListFull(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		httperr.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

// Handler unico para upload de laudo
// POST /:patientID/labs/upload
// field: file (PDF/JPEG/PNG)
func (h *LabsHandler) UploadAndProcessLabs(c *gin.Context) {
	user := middleware.MustGetCurrentUser(c)

	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	documentURI, mimeType, err := h.handleFileUpload(c, patientID)
	if err != nil {
		httperr.WriteError(c, err)
		return
	}

	output, err := h.createUC.Execute(c.Request.Context(), labsuc.CreateLabReportFromDocumentInput{
		PatientID:        patientID,
		DocumentURI:      documentURI,
		MimeType:         mimeType,
		UploadedByUserID: user.ID,
	})
	if err != nil {
		httperr.WriteError(c, err)
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
	const MaxFileSize = 15 * 1024 * 1024 // 15MB

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return "", "", &apperr.AppError{
			Code:    apperr.REQUIRED_FIELD_MISSING,
			Message: "arquivo é obrigatório",
			Cause:   err,
		}
	}
	if fileHeader.Size == 0 {
		return "", "", &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "arquivo vazio",
		}
	}

	if fileHeader.Size > MaxFileSize {
		return "", "", &apperr.AppError{
			Code:    apperr.UPLOAD_SIZE_EXCEEDED,
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

	if !isSupportedMimeType(contentType) {
		return "", "", &apperr.AppError{
			Code:    apperr.INVALID_FIELD_FORMAT,
			Message: "tipo de arquivo não suportado",
			Cause:   fmt.Errorf("content_type=%s", contentType),
		}
	}

	uniqueID := uuid.NewString()
	ext := mimeToExt(contentType)
	if ext == "" {
		return "", "", &apperr.AppError{
			Code:    apperr.INVALID_FIELD_FORMAT,
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
			Code:    apperr.INFRA_STORAGE_ERROR,
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
			httperr.WriteError(c, &apperr.AppError{
				Code:    apperr.VALIDATION_FAILED,
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
			httperr.WriteError(c, &apperr.AppError{
				Code:    apperr.VALIDATION_FAILED,
				Message: "offset deve ser >= 0",
				Cause:   err,
			})
			return 0, 0, false
		}
		offset = o
	}

	return limit, offset, true
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
