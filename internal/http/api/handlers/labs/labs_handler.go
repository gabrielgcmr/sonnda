package labs

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

	labsvc "sonnda-api/internal/app/services/labs"

	"sonnda-api/internal/domain/ports/integrations"
	"sonnda-api/internal/http/api/handlers/common"
	"sonnda-api/internal/http/middleware"

	applog "sonnda-api/internal/app/observability"
)

type LabsHandler struct {
	svc     labsvc.Service
	storage integrations.StorageService
}

func NewLabsHandler(
	svc labsvc.Service,
	storageClient integrations.StorageService,
) *LabsHandler {
	return &LabsHandler{
		svc:     svc,
		storage: storageClient,
	}
}

func (h *LabsHandler) ListLabs(c *gin.Context) {
	if h == nil || h.svc == nil {
		common.RespondError(c, http.StatusInternalServerError, "labs_not_configured", errors.New("labs service not configured"))
		return
	}

	patientID := c.Param("id")
	if patientID == "" {
		patientID = c.Param("patientID")
	}
	if patientID == "" {
		common.RespondError(c, http.StatusBadRequest, "missing_patient_id", nil)
		return
	}

	parsedID, err := uuid.Parse(patientID)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}

	list, err := h.svc.List(c.Request.Context(), parsedID, limit, offset)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *LabsHandler) ListFullLabs(c *gin.Context) {
	if h == nil || h.svc == nil {
		common.RespondError(c, http.StatusInternalServerError, "labs_not_configured", errors.New("labs service not configured"))
		return
	}

	patientID := c.Param("id")
	if patientID == "" {
		patientID = c.Param("patientID")
	}
	if patientID == "" {
		common.RespondError(c, http.StatusBadRequest, "missing_patient_id", nil)
		return
	}

	parsedID, err := uuid.Parse(patientID)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}

	list, err := h.svc.ListFull(c.Request.Context(), parsedID, limit, offset)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

// Handler unico para upload de laudo
// POST /:patientID/labs/upload
// field: file (PDF/JPEG/PNG)
func (h *LabsHandler) UploadAndProcessLabs(c *gin.Context) {
	if h == nil || h.svc == nil || h.storage == nil {
		common.RespondError(c, http.StatusInternalServerError, "labs_not_configured", errors.New("labs dependencies not configured"))
		return
	}

	log := applog.FromContext(c.Request.Context())

	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		return
	}

	patientID := c.Param("id")
	if patientID == "" {
		patientID = c.Param("patientID")
	}
	if patientID == "" {
		common.RespondError(c, http.StatusBadRequest, "missing_patient_id", nil)
		return
	}

	parsedID, err := uuid.Parse(patientID)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	documentURI, mimeType, err := h.handleFileUpload(c)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "upload_failed", err)
		return
	}

	output, err := h.svc.CreateFromDocument(c.Request.Context(), labsvc.CreateFromDocumentInput{
		PatientID:        parsedID,
		DocumentURI:      documentURI,
		MimeType:         mimeType,
		UploadedByUserID: user.ID,
	})
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	log.Info("labs_report_created", slog.String("patient_id", patientID))
	c.JSON(http.StatusCreated, output)
}

// handleFileUpload centraliza toda a logica de:
// - ler o arquivo do multipart
// - detectar/validar content-type
// - fazer upload pro storage
// - retornar (URI, MIME)
func (h *LabsHandler) handleFileUpload(
	c *gin.Context,
) (string, string, error) {
	const MaxFileSize = 15 * 1024 * 1024 // 15MB

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return "", "", fmt.Errorf("file_required: %w", err)
	}
	if fileHeader.Size == 0 {
		return "", "", fmt.Errorf("empty_file")
	}

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

	uniqueID := uuid.NewString()
	ext := mimeToExt(contentType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported_mime_type:%s", contentType)
	}

	patientIDStr := c.Param("id")
	if patientIDStr == "" {
		patientIDStr = c.Param("patientID")
	}
	if patientIDStr == "" {
		return "", "", fmt.Errorf("missing_patient_id")
	}

	objectName := fmt.Sprintf("patients/%s/lab-reports/%s%s", patientIDStr, uniqueID, ext)

	uri, err := h.storage.Upload(c.Request.Context(), file, objectName, contentType)
	if err != nil {
		return "", "", fmt.Errorf("upload_failed: %w", err)
	}

	return uri, contentType, nil
}

func parsePagination(c *gin.Context, defaultLimit, defaultOffset int) (limit, offset int, ok bool) {
	limit = defaultLimit
	offset = defaultOffset

	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			common.RespondError(c, http.StatusBadRequest, "invalid_limit", errors.New("limit must be > 0"))
			return 0, 0, false
		}
		limit = l
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			common.RespondError(c, http.StatusBadRequest, "invalid_offset", errors.New("offset must be >= 0"))
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
