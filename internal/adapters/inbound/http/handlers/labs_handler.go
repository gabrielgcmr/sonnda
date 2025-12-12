package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/services"
	"sonnda-api/internal/core/usecases/labs"

	"github.com/gin-gonic/gin"
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
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	// 2) Paciente alvo (dono do laudo)
	patientID := c.Param("patientID")
	if patientID == "" {
		// Se suas rotas usam :id em vez de :patientID, você pode dar fallback:
		patientID = c.Param("id")
	}
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_patient_id"})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = l
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_offset"})
			return
		}
		offset = o
	}

	list, err := h.listLabs.Execute(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "patient_not_found",
				"message": "nenhum paciente vinculado a este usuário",
			})
		case domain.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "usuário não permitido para esta operação",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, list)

}

// handleFileUpload centraliza toda a lógica de:
// - ler o arquivo do multipart
// - detectar/validar content-type
// - fazer upload pro storage
// - retornar (URI, MIME)
func (h *LabsHandler) handleFileUpload(
	c *gin.Context,
	patientID string,
) (string, string, error) {

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return "", "", fmt.Errorf("file_required: %w", err)
	}

	if fileHeader.Size == 0 {
		return "", "", fmt.Errorf("empty_file")
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

	objectName := fmt.Sprintf("patients/%s/lab-reports/%s", patientID, fileHeader.Filename)

	uri, err := h.storage.Upload(c.Request.Context(), file, objectName, contentType)
	if err != nil {
		return "", "", fmt.Errorf("upload_failed: %w", err)
	}

	return uri, contentType, nil
}

func (h *LabsHandler) ListFullLabs(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	// 2) Paciente alvo (dono do laudo)
	patientID := c.Param("patientID")
	if patientID == "" {
		// fallback para rotas que usam :id
		patientID = c.Param("id")
	}
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_patient_id"})
		return
	}

	// paginação (mesma lógica do ListLabs)
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = l
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_offset"})
			return
		}
		offset = o
	}

	// aqui chamamos o usecase de lista COMPLETA
	list, err := h.listFullLabs.Execute(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "patient_not_found",
				"message": "nenhum paciente vinculado a este usuário",
			})
		case domain.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "usuário não permitido para esta operação",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, list)
}

// Handler único para upload de laudo
// POST /:patientID/labs/upload
// field: file (PDF/JPEG/PNG)
func (h *LabsHandler) UploadAndProcessLabs(c *gin.Context) {

	// 1) Usuário autenticado (quem está fazendo o upload)
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuario não autenticado.",
		})
		return
	}

	// 2) Paciente alvo (dono do laudo)
	patientID := c.Param("patientID")
	if patientID == "" {
		// Se suas rotas usam :id em vez de :patientID, você pode dar fallback:
		patientID = c.Param("id")
	}
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_patient_id"})
		return
	}

	// 🔐 Aqui é um bom lugar pra checar autorização:
	// if !h.authorization.CanUploadLab(ctx, user, patientID) { ... }

	// 3) Centraliza toda a lógica de upload (arquivo, mimetype, GCS, etc.)
	documentURI, mimeType, err := h.handleFileUpload(c, patientID)
	if err != nil {
		// você pode melhorar o parsing dessa error message
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "upload_failed",
			"details": err.Error(),
		})
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
		switch err {
		case domain.ErrInvalidInput, domain.ErrInvalidDocument:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_document",
				"details": err.Error(),
			})
		case domain.ErrDocumentProcessing:
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "document_processing_failed",
				"details": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "processing_failed",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, output)
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
