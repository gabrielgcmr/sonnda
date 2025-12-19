package handlers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/ports/services"
	"sonnda-api/internal/core/usecases/labs"
	applog "sonnda-api/internal/logger"
)

type LabsHandler struct {
	createUC     *labs.ExtractLabsUseCase
	listLabs     *labs.ListLabsUseCase
	listFullLabs *labs.ListFullLabsUseCase
	storage      services.StorageService
}

func NewLabsHandler(
	createUC *labs.ExtractLabsUseCase,
	listLabs *labs.ListLabsUseCase,
	listFullLabs *labs.ListFullLabsUseCase,
	storageClient services.StorageService,
) *LabsHandler {
	return &LabsHandler{
		createUC:     createUC,
		listLabs:     listLabs,
		listFullLabs: listFullLabs,
		storage:      storageClient,
	}
}

func (h *LabsHandler) ListLabs(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	// 1) Paciente alvo (dono do laudo)

	patientID, ok := parsePatientID(c, log)
	if !ok {
		return
	}

	// 2) paginação
	limit, offset, ok := parsePagination(c, log, 100, 0)
	if !ok {
		return
	}

	list, err := h.listLabs.Execute(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, list)

}

func (h *LabsHandler) ListFullLabs(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	// 1) Paciente alvo (dono do laudo)
	patientID, ok := parsePatientID(c, log)
	if !ok {
		return
	}

	// 2. paginação (mesma lógica do ListLabs)
	limit, offset, ok := parsePagination(c, log, 100, 0)
	if !ok {
		return
	}

	// 3. aqui chamamos o usecase de lista COMPLETA
	list, err := h.listFullLabs.Execute(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

// Handler único para upload de laudo
// POST /:patientID/labs/upload
// field: file (PDF/JPEG/PNG)
func (h *LabsHandler) UploadAndProcessLabs(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	// 1) Paciente alvo (dono do laudo)
	patientID, ok := parsePatientID(c, log)
	if !ok {
		return
	}

	// 2) Centraliza toda a lógica de upload (arquivo, mimetype, GCS, etc.)
	documentURI, mimeType, err := h.handleFileUpload(c)
	if err != nil {
		RespondUploadError(c, log, err)
		return
	}

	output, err := h.createUC.Execute(
		c.Request.Context(),
		labs.CreateFromDocumentInput{
			PatientID:        patientID,
			DocumentURI:      documentURI,
			MimeType:         mimeType,
			UploadedByUserID: user.ID,
		})
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	log.Info("labs_report_created", slog.String("patient_id", patientID.String()))
	c.JSON(http.StatusCreated, output)
}

// handleFileUpload centraliza toda a lógica de:
// - ler o arquivo do multipart
// - detectar/validar content-type
// - fazer upload pro storage
// - retornar (URI, MIME)
func (h *LabsHandler) handleFileUpload(
	c *gin.Context,
) (string, string, error) {
	const MaxFileSize = 15 * 1024 * 1024 // 15MB

	// Validações

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return "", "", fmt.Errorf("file_required: %w", err)
	}
	if fileHeader.Size == 0 {
		return "", "", fmt.Errorf("empty_file")
	}

	// Validação de Tamanho
	if fileHeader.Size > MaxFileSize {
		return "", "", fmt.Errorf("file_too_large")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("open_file_failed: %w", err)
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
		return "", "", fmt.Errorf("unsupported_mime_type: %s", contentType)
	}

	uniqueID := uuid.New().String()
	ext := mimeToExt(contentType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported_mime_type:%s", contentType)
	}

	// Obtain and validate patient ID from route parameters to organize storage per patient.
	patientIDStr := c.Param("patientID")
	if patientIDStr == "" {
		patientIDStr = c.Param("id")
	}
	if patientIDStr == "" {
		return "", "", fmt.Errorf("missing_patient_id")
	}
	if _, err := uuid.Parse(patientIDStr); err != nil {
		return "", "", fmt.Errorf("invalid_patient_id: %w", err)
	}
	// Store as: patients/{patient-id}/lab-reports/{unique-id}{ext}
	objectName := fmt.Sprintf("patients/%s/lab-reports/%s%s", patientIDStr, uniqueID, ext)

	uri, err := h.storage.Upload(c.Request.Context(), file, objectName, contentType)
	if err != nil {
		return "", "", fmt.Errorf("upload_failed: %w", err)
	}

	return uri, contentType, nil
}

func parsePatientID(c *gin.Context, log *slog.Logger) (uuid.UUID, bool) {
	idStr := c.Param("patientID")
	if idStr == "" {
		idStr = c.Param("id")
	}
	if idStr == "" {
		RespondError(c, log, http.StatusBadRequest, "missing_patient_id", nil)
		return uuid.Nil, false
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_patient_id", err)
		return uuid.Nil, false
	}

	return id, true
}

func parsePagination(c *gin.Context, log *slog.Logger, defaultLimit, defaultOffset int) (limit, offset int, ok bool) {
	limit = defaultLimit
	offset = defaultOffset

	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			RespondError(c, log, http.StatusBadRequest, "invalid_limit", errors.New("limit must be > 0"))
			return 0, 0, false
		}
		limit = l
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			RespondError(c, log, http.StatusBadRequest, "invalid_offset", errors.New("offset must be >= 0"))
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
