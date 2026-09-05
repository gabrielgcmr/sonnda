// internal/api/handlers/exams.go
package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/rbac"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamsHandler struct {
	svc           examsvc.Service
	createLabUC   labsuc.CreateLabReportFromDocumentUseCase
	storage       domainstorage.FileStorageService
	textExtractor domaintext.Extractor
	authz         authorization.Authorizer
}

func NewExams(
	svc examsvc.Service,
	createLabUC labsuc.CreateLabReportFromDocumentUseCase,
	storageClient domainstorage.FileStorageService,
	textExtractor domaintext.Extractor,
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

func (h *ExamsHandler) ListExamDocumentTexts(c *gin.Context) {
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

	list, err := h.svc.ListDocumentTextsByPatient(c.Request.Context(), patientID, limit, offset)
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

	routed := h.extractAndRoute(c, output.ID, upload)
	if routed == nil {
		routed = h.routeFromMetadata(c, output.ID)
	}
	if routed != nil {
		output = routed.document
		if h.isLaboratoryDocument(output) {
			labReport, err := h.createLabReportIfNeeded(c, output, upload)
			if err != nil {
				message := "falha no processamento laboratorial"
				var appErr *apperr.AppError
				if errors.As(err, &appErr) {
					message = appErr.Message
				}
				// O upload fica disponivel, mas nao aparenta ter sido processado.
				if _, markErr := h.svc.MarkFailed(c.Request.Context(), examsvc.MarkExamDocumentFailedInput{
					ID: output.ID, ErrorMessage: message,
				}); markErr != nil {
					err = apperr.Internal("falha ao registrar erro do exame", errors.Join(err, markErr))
				}
				presenter.ErrorResponder(c, err)
				return
			}
			if labReport != nil {
				h.createExamDocumentTextFromLab(c, output, labReport)
			}
		} else {
			h.createExamDocumentTextIfPossible(c, output, routed.extracted)
		}
	}

	c.JSON(http.StatusCreated, output)
}

func (h *ExamsHandler) createExamDocumentTextIfPossible(c *gin.Context, document *examsvc.ExamDocumentOutput, extracted *domaintext.ExtractOutput) {
	if document == nil || extracted == nil {
		return
	}

	category := exams.ExamTypeUnknown
	if document.ExamType != nil {
		category = *document.ExamType
	}

	_, err := h.svc.CreateDocumentTextFromText(c.Request.Context(), examsvc.CreateExamDocumentTextFromTextInput{
		ExamDocumentID:   document.ID,
		PatientID:        document.PatientID,
		UploadedByUserID: document.UploadedByUserID,
		Category:         category,
		Text:             extracted.Text,
		ExtractionMethod: extracted.Method,
		Confidence:       document.Confidence,
	})
	if err != nil {
		// O texto fica em exam_documents; retry pode criar exam_document_texts depois.
		return
	}
}

func (h *ExamsHandler) createExamDocumentTextFromLab(c *gin.Context, document *examsvc.ExamDocumentOutput, labReport *labsvc.LabReportOutput) {
	if document == nil || labReport == nil {
		return
	}

	text := buildLabReportText(labReport)
	if strings.TrimSpace(text) == "" {
		return
	}

	method := "lab_document_ai"
	_, err := h.svc.CreateDocumentTextFromText(c.Request.Context(), examsvc.CreateExamDocumentTextFromTextInput{
		ExamDocumentID:   document.ID,
		PatientID:        document.PatientID,
		UploadedByUserID: document.UploadedByUserID,
		Category:         exams.ExamTypeLaboratory,
		Text:             text,
		ExtractionMethod: method,
		Confidence:       document.Confidence,
	})
	if err != nil {
		// Labs estruturado ja foi salvo; texto generico pode ser recriado depois.
		return
	}
}

func (h *ExamsHandler) createLabReportIfNeeded(c *gin.Context, document *examsvc.ExamDocumentOutput, upload *uploadedExamFile) (*labsvc.LabReportOutput, error) {
	if document == nil || !h.isLaboratoryDocument(document) {
		return nil, nil
	}
	if h.createLabUC == nil {
		return nil, apperr.Internal("processamento laboratorial indisponivel", nil)
	}

	return h.createLabUC.Execute(c.Request.Context(), labsuc.CreateLabReportFromDocumentInput{
		PatientID:        document.PatientID,
		ExamDocumentID:   &document.ID,
		DocumentURI:      upload.storageURI,
		MimeType:         upload.mimeType,
		UploadedByUserID: document.UploadedByUserID,
	})
}

func (h *ExamsHandler) isLaboratoryDocument(document *examsvc.ExamDocumentOutput) bool {
	return document != nil && document.ExamType != nil && *document.ExamType == exams.ExamTypeLaboratory
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
	extracted *domaintext.ExtractOutput
}

func (h *ExamsHandler) extractAndRoute(c *gin.Context, documentID uuid.UUID, upload *uploadedExamFile) *routedExamDocument {
	if h.textExtractor == nil {
		return nil
	}

	extracted, err := h.textExtractor.Extract(c.Request.Context(), domaintext.ExtractInput{
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

func (h *ExamsHandler) routeFromMetadata(c *gin.Context, documentID uuid.UUID) *routedExamDocument {
	output, err := h.svc.RouteDocument(c.Request.Context(), examsvc.RouteExamDocumentInput{
		ID:               documentID,
		ExtractionMethod: "metadata",
	})
	if err != nil {
		// O upload ja existe; classificacao pode ser refeita depois.
		return nil
	}
	if output == nil {
		return nil
	}
	return &routedExamDocument{document: output}
}

func buildLabReportText(report *labsvc.LabReportOutput) string {
	if report == nil {
		return ""
	}

	var builder strings.Builder
	writeLine := func(parts ...string) {
		line := strings.TrimSpace(strings.Join(parts, " "))
		if line == "" {
			return
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	writeLine("Exame laboratorial")
	writeOptionalLine(&builder, "Paciente", report.PatientName)
	writeOptionalLine(&builder, "Laboratorio", report.LabName)
	writeOptionalLine(&builder, "Solicitante", report.RequestingDoctor)
	writeDateLine(&builder, "Data do laudo", report.ReportDate)

	for _, result := range report.TestResults {
		builder.WriteByte('\n')
		writeLine(result.TestName)
		writeOptionalLine(&builder, "Material", result.Material)
		writeOptionalLine(&builder, "Metodo", result.Method)
		writeDateLine(&builder, "Coletado em", result.CollectedAt)
		writeDateLine(&builder, "Liberado em", result.ReleaseAt)

		for _, item := range result.Items {
			writeLine(formatLabItem(item))
		}
	}

	return strings.TrimSpace(builder.String())
}

func writeOptionalLine(builder *strings.Builder, label string, value *string) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(strings.TrimSpace(*value))
	builder.WriteByte('\n')
}

func writeDateLine(builder *strings.Builder, label string, value *time.Time) {
	if value == nil {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(value.Format("02/01/2006"))
	builder.WriteByte('\n')
}

func formatLabItem(item labsvc.TestItemOutput) string {
	parts := []string{"-", strings.TrimSpace(item.ParameterName)}
	if item.ResultValue != nil && strings.TrimSpace(*item.ResultValue) != "" {
		parts = append(parts, strings.TrimSpace(*item.ResultValue))
	}
	if item.ResultUnit != nil && strings.TrimSpace(*item.ResultUnit) != "" {
		parts = append(parts, strings.TrimSpace(*item.ResultUnit))
	}
	if item.ReferenceText != nil && strings.TrimSpace(*item.ReferenceText) != "" {
		parts = append(parts, "(Referencia:", strings.TrimSpace(*item.ReferenceText)+")")
	}
	return strings.Join(parts, " ")
}
