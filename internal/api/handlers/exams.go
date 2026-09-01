package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	domaindoc "github.com/gabrielgcmr/sonnda/internal/domain/documenttext"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/rbac"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamsHandler struct {
	svc           examsvc.Service
	createLabUC   labsuc.CreateLabReportFromDocumentUseCase
	storage       domainstorage.FileStorageService
	textExtractor domaindoc.Extractor
	authz         authorization.Authorizer
}

func NewExams(
	svc examsvc.Service,
	createLabUC labsuc.CreateLabReportFromDocumentUseCase,
	storageClient domainstorage.FileStorageService,
	textExtractor domaindoc.Extractor,
	authz authorization.Authorizer,
) *ExamsHandler {
	return &ExamsHandler{
		svc:           svc,
		createLabUC:   createLabUC,
		storage:       storageClient,
		textExtractor: textExtractor,
		authz:         authz,
	}
}

func (h *ExamsHandler) ListExamDocuments(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	if h.authz != nil {
		if err := h.authz.Require(c.Request.Context(), currentUser, rbac.ActionReadExams, &patientID); err != nil {
			presenter.ErrorResponder(c, err)
			return
		}
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

func (h *ExamsHandler) ListExamReports(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	patientID, ok := parsePatientIDParam(c, "id")
	if !ok {
		return
	}

	if h.authz != nil {
		if err := h.authz.Require(c.Request.Context(), currentUser, rbac.ActionReadExams, &patientID); err != nil {
			presenter.ErrorResponder(c, err)
			return
		}
	}

	limit, offset, ok := parsePagination(c, 100, 0)
	if !ok {
		return
	}

	list, err := h.svc.ListReportsByPatient(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
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
	defer os.Remove(upload.localPath)

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

	if h.textExtractor != nil {
		if routed := h.extractAndRoute(c, output.ID, upload); routed != nil {
			output = routed.document
			h.createExamReportIfPossible(c, output, routed.extracted)
			h.createLabReportIfNeeded(c, output, upload)
		}
	}

	c.JSON(http.StatusCreated, output)
}

func (h *ExamsHandler) createExamReportIfPossible(c *gin.Context, document *examsvc.ExamDocumentOutput, extracted *domaindoc.ExtractOutput) {
	if document == nil || extracted == nil {
		return
	}

	category := exams.ExamTypeUnknown
	if document.ExamType != nil {
		category = *document.ExamType
	}

	_, err := h.svc.CreateReportFromText(c.Request.Context(), examsvc.CreateExamReportFromTextInput{
		ExamDocumentID:   document.ID,
		PatientID:        document.PatientID,
		UploadedByUserID: document.UploadedByUserID,
		Category:         category,
		ReportText:       extracted.Text,
		ExtractionMethod: extracted.Method,
		Confidence:       document.Confidence,
	})
	if err != nil {
		// O texto fica em exam_documents; retry pode criar exam_reports depois.
		return
	}
}

func (h *ExamsHandler) createLabReportIfNeeded(c *gin.Context, document *examsvc.ExamDocumentOutput, upload *uploadedExamFile) {
	if h.createLabUC == nil || document == nil || document.ExamType == nil {
		return
	}
	if *document.ExamType != exams.ExamTypeLaboratory {
		return
	}

	_, err := h.createLabUC.Execute(c.Request.Context(), labsuc.CreateLabReportFromDocumentInput{
		PatientID:        document.PatientID,
		ExamDocumentID:   &document.ID,
		DocumentURI:      upload.storageURI,
		MimeType:         upload.mimeType,
		UploadedByUserID: document.UploadedByUserID,
	})
	if err != nil {
		// Laudo textual ja existe; labs estruturado pode ser reprocessado depois.
		return
	}
}

type uploadedExamFile struct {
	storageURI       string
	originalFilename string
	mimeType         string
	localPath        string
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

	extHint := filepath.Ext(fileHeader.Filename)
	tempFile, err := os.CreateTemp("", "sonnda-exam-*"+extHint)
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

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		detectFile, err := os.Open(tempPath)
		if err != nil {
			return nil, apperr.Internal("falha ao detectar tipo de arquivo", err)
		}
		buf := make([]byte, 512)
		n, _ := detectFile.Read(buf)
		_ = detectFile.Close()
		contentType = http.DetectContentType(buf[:n])
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
	uploadFile, err := os.Open(tempPath)
	if err != nil {
		return nil, apperr.Internal("falha ao reabrir arquivo", err)
	}
	defer uploadFile.Close()

	uri, err := h.storage.Upload(c.Request.Context(), uploadFile, objectName, contentType)
	if err != nil {
		return nil, &apperr.AppError{
			Kind:    apperr.INFRA_STORAGE_ERROR,
			Message: "falha no upload",
			Cause:   err,
		}
	}

	removeTempOnError = false
	return &uploadedExamFile{
		storageURI:       uri,
		originalFilename: fileHeader.Filename,
		mimeType:         contentType,
		localPath:        tempPath,
	}, nil
}

type routedExamDocument struct {
	document  *examsvc.ExamDocumentOutput
	extracted *domaindoc.ExtractOutput
}

func (h *ExamsHandler) extractAndRoute(c *gin.Context, documentID uuid.UUID, upload *uploadedExamFile) *routedExamDocument {
	extracted, err := h.textExtractor.Extract(c.Request.Context(), domaindoc.ExtractInput{
		LocalPath:        upload.localPath,
		MimeType:         upload.mimeType,
		OriginalFilename: upload.originalFilename,
	})
	if err != nil {
		// Sem fallback caro nesta etapa; o documento fica como uploaded.
		return nil
	}

	output, err := h.svc.RouteDocument(c.Request.Context(), examsvc.RouteExamDocumentInput{
		ID:               documentID,
		ExtractedText:    extracted.Text,
		ExtractionMethod: extracted.Method,
	})
	if err != nil {
		// Upload ja foi salvo; reprocessamento pode ocorrer depois.
		return nil
	}

	return &routedExamDocument{
		document:  output,
		extracted: extracted,
	}
}
