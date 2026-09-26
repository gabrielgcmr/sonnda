// internal/api/handlers/exams.go
package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	documenthttp "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/http"
	processing "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamsHandler struct {
	svc           documents.Service
	processor     processing.ProcessStoredDocumentUseCase
	storage       domainstorage.FileStorageService
	accessChecker patientaccess.Checker
}

func NewExams(
	svc documents.Service,
	processor processing.ProcessStoredDocumentUseCase,
	storageClient domainstorage.FileStorageService,
	accessChecker patientaccess.Checker,
) *ExamsHandler {
	return &ExamsHandler{svc: svc, processor: processor, storage: storageClient, accessChecker: accessChecker}
}

func (h *ExamsHandler) ListExamDocuments(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	patientID, ok := parsePatientIDParam(c, "patientId")
	if !ok {
		return
	}
	if err := h.accessChecker.RequireAccess(c.Request.Context(), currentUser.ID, patientID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}
	list, err := h.svc.ListByPatient(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ExamsHandler) GetExamDocument(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	documentID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		presenter.ErrorResponder(c, apperr.Validation("document_id inválido", apperr.Violation{Field: "document_id", Reason: "invalid"}))
		return
	}
	document, err := h.svc.FindByID(c.Request.Context(), documentID)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	if document == nil {
		presenter.ErrorResponder(c, apperr.NotFound("documento nao encontrado"))
		return
	}
	if err := h.accessChecker.RequireAccess(c.Request.Context(), currentUser.ID, document.PatientID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	c.JSON(http.StatusOK, document)
}

func (h *ExamsHandler) ListExamDocumentTexts(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	patientID, ok := parsePatientIDParam(c, "patientId")
	if !ok {
		return
	}
	if err := h.accessChecker.RequireAccess(c.Request.Context(), currentUser.ID, patientID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}
	list, err := h.svc.ListDocumentTextsByPatient(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ExamsHandler) UploadExamDocument(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	patientID, ok := parsePatientIDParam(c, "patientId")
	if !ok {
		return
	}
	if err := h.accessChecker.RequireAccess(c.Request.Context(), currentUser.ID, patientID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	collectionDate, err := parseExamCollectionDate(c.PostForm("collection_date"))
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{Kind: apperr.REQUIRED_FIELD_MISSING, Message: "arquivo e obrigatorio", Cause: err})
		return
	}
	upload, err := documenthttp.UploadDocument(c.Request.Context(), fileHeader, patientID, h.storage)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	defer os.Remove(upload.LocalPath)

	document, err := h.processor.Execute(c.Request.Context(), processing.ProcessStoredDocumentInput{
		PatientID:        patientID,
		UploadedByUserID: currentUser.ID,
		StorageURI:       upload.StorageURI,
		OriginalFilename: upload.OriginalFilename,
		MimeType:         upload.MimeType,
		LocalPath:        upload.LocalPath,
		CollectionDate:   collectionDate,
	})
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	c.JSON(http.StatusCreated, document)
}

func parseExamCollectionDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apperr.Validation("data da coleta invalida", apperr.Violation{Field: "collection_date", Reason: "must_be_yyyy_mm_dd"})
	}
	return &date, nil
}
