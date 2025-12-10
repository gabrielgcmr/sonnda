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

func (h *LabsHandler) ListMyLabReports(c *gin.Context) {
	//todo: implementar paginação
}

// UploadAndProcessLabReport
// POST /patients/:patientID/lab-reports/upload
// Content-Type: multipart/form-data
// field: file (PDF/JPEG/PNG)
func (h *LabsHandler) UploadAndProcessLabReport(c *gin.Context) {
	ctx := c.Request.Context()

	patientID := c.Param("patientID")
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_patient_id"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_required", "details": err.Error()})
		return
	}

	if fileHeader.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty_file"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open_file_failed", "details": err.Error()})
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		detected := http.DetectContentType(buf[:n])

		if seeker, ok := file.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}

		contentType = detected
	}

	if !isSupportedMimeType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "unsupported_mime_type",
			"content_type":  contentType,
			"allowed_types": []string{"application/pdf", "image/jpeg", "image/png"},
		})
		return
	}

	objectName := fmt.Sprintf("patients/%s/lab-reports/%s", patientID, fileHeader.Filename)

	gcsURI, err := h.storage.Upload(ctx, file, objectName, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload_failed", "details": err.Error()})
		return
	}

	report, err := h.createUC.Execute(ctx, labs.CreateFromDocumentInput{
		PatientID:   patientID,
		DocumentURI: gcsURI,
		MimeType:    contentType,
	})
	if err != nil {
		switch err {
		case domain.ErrInvalidInput, domain.ErrInvalidDocument:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_document", "details": err.Error()})
		case domain.ErrDocumentProcessing:
			c.JSON(http.StatusBadGateway, gin.H{"error": "document_processing_failed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "processing_failed", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, report)
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

func (h *LabsHandler) UploadAndProcessMyLabReport(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// aqui você lê o arquivo / gera DocumentURI / MimeType como já faz no fluxo do médico
	documentURI, mimeType, err := h.extractUploadInfo(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_upload"})
		return
	}

	output, err := h.CreateFromDocumentForCurrentUserInput.Execute(
		c.Request.Context(),
		user,
		labs.CreateFromDocumentForCurrentUserInput{
			DocumentURI: documentURI,
			MimeType:    mimeType,
		},
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}
