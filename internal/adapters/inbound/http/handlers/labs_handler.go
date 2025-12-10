package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/services"
	"sonnda-api/internal/core/usecases/labs"

	"github.com/gin-gonic/gin"
)

type LabsHandler struct {
	createUC *labs.CreateFromDocumentUseCase
	storage  services.StorageService
}

func NewLabsHandler(
	createUC *labs.CreateFromDocumentUseCase,
	storageClient services.StorageService,
) *LabsHandler {
	return &LabsHandler{
		createUC: createUC,
		storage:  storageClient,
	}
}

func (h *LabsHandler) ListLabReports(c *gin.Context) {
	//todo: implementar paginação
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

// Handler único para upload de laudo
// POST /patients/:patientID/labs/upload
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
				"error": "document_processing_failed",
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
