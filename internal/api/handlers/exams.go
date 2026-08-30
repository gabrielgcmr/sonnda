package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/rbac"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamsHandler struct {
	svc     examsvc.Service
	storage domainstorage.FileStorageService
	authz   authorization.Authorizer
}

func NewExams(
	svc examsvc.Service,
	storageClient domainstorage.FileStorageService,
	authz authorization.Authorizer,
) *ExamsHandler {
	return &ExamsHandler{
		svc:     svc,
		storage: storageClient,
		authz:   authz,
	}
}

// UploadExamDocument saves the original exam document for later routing.
// POST /v1/patients/:id/exames
// field: file (PDF/JPEG/PNG)
func (h *ExamsHandler) UploadExamDocument(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	if h.authz != nil {
		if err := h.authz.Require(c.Request.Context(), currentUser, rbac.ActionUploadExams, &patientID); err != nil {
			presenter.ErrorResponder(c, err)
			return
		}
	}

	upload, err := h.handleExamFileUpload(c, patientID)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	output, err := h.svc.Create(c.Request.Context(), examsvc.CreateExamDocumentInput{
		PatientID:        patientID,
		UploadedByUserID: currentUser.ID,
		StorageURI:       upload.storageURI,
		OriginalFilename: upload.originalFilename,
		MimeType:         upload.mimeType,
	})
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

type uploadedExamFile struct {
	storageURI       string
	originalFilename string
	mimeType         string
}

func (h *ExamsHandler) handleExamFileUpload(c *gin.Context, patientID uuid.UUID) (*uploadedExamFile, error) {
	const maxFileSize = 10 * 1024 * 1024 // 10MB

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, &apperr.AppError{
			Kind:    apperr.REQUIRED_FIELD_MISSING,
			Message: "arquivo e obrigatorio",
			Cause:   err,
		}
	}
	if fileHeader.Size == 0 {
		return nil, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "arquivo vazio",
		}
	}
	if fileHeader.Size > maxFileSize {
		return nil, &apperr.AppError{
			Kind:    apperr.UPLOAD_SIZE_EXCEEDED,
			Message: "arquivo muito grande",
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apperr.Internal("falha ao abrir arquivo", err)
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
		return nil, &apperr.AppError{
			Kind:    apperr.INVALID_FIELD_FORMAT,
			Message: "tipo de arquivo nao suportado",
			Cause:   fmt.Errorf("content_type=%s", contentType),
		}
	}

	ext := mimeToExt(contentType)
	if ext == "" {
		return nil, &apperr.AppError{
			Kind:    apperr.INVALID_FIELD_FORMAT,
			Message: "tipo de arquivo nao suportado",
			Cause:   fmt.Errorf("content_type=%s", contentType),
		}
	}
	if patientID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}

	objectName := fmt.Sprintf("patients/%s/exam-documents/%s%s", patientID.String(), uuid.NewString(), ext)
	uri, err := h.storage.Upload(c.Request.Context(), file, objectName, contentType)
	if err != nil {
		return nil, &apperr.AppError{
			Kind:    apperr.INFRA_STORAGE_ERROR,
			Message: "falha no upload",
			Cause:   err,
		}
	}

	return &uploadedExamFile{
		storageURI:       uri,
		originalFilename: fileHeader.Filename,
		mimeType:         contentType,
	}, nil
}
