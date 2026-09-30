// internal/features/documentprocessing/http/exams.go
package http

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/humaerror"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	processing "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type ExamsHandler struct {
	svc           documents.Service
	processor     processing.ProcessStoredDocumentUseCase
	storage       domainstorage.FileStorageService
	accessChecker patientaccess.Checker
}

const examDocumentFileURLExpirationMinutes = 15

type examDocumentFileResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type listExamDocumentsInput struct {
	PatientID uuid.UUID `path:"patientId" format:"uuid"`
	Limit     int       `query:"limit" default:"100" minimum:"1" maximum:"100"`
	Offset    int       `query:"offset" default:"0" minimum:"0"`
}

type examDocumentInput struct {
	DocumentID uuid.UUID `path:"documentId" format:"uuid"`
}

type listExamDocumentsOutput struct {
	Body []documents.ExamDocumentOutput
}

type listExamDocumentTextsOutput struct {
	Body []documents.ExamDocumentTextOutput
}

type examDocumentOutput struct {
	Body documents.ExamDocumentOutput
}

type examDocumentFileOutput struct {
	Body examDocumentFileResponse
}

type uploadExamDocumentForm struct {
	File           huma.FormFile `form:"file" required:"true"`
	CollectionDate string        `form:"collection_date"`
}

type uploadExamDocumentInput struct {
	PatientID uuid.UUID `path:"patientId" format:"uuid"`
	RawBody   huma.MultipartFormFiles[uploadExamDocumentForm]
}

func NewExams(
	svc documents.Service,
	processor processing.ProcessStoredDocumentUseCase,
	storageClient domainstorage.FileStorageService,
	accessChecker patientaccess.Checker,
) *ExamsHandler {
	return &ExamsHandler{svc: svc, processor: processor, storage: storageClient, accessChecker: accessChecker}
}

// RegisterHumaRoutes registers the supported exam-document operations.
func (h *ExamsHandler) RegisterHumaRoutes(registered huma.API, security []map[string][]string) {
	huma.Register(registered, huma.Operation{
		OperationID: "listExamDocuments",
		Method:      http.MethodGet,
		Path:        "/patients/{patientId}/exam-documents",
		Summary:     "Listar documentos de exame",
		Tags:        []string{"Exam documents"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
		Security:    security,
	}, h.listExamDocuments)

	huma.Register(registered, huma.Operation{
		OperationID: "listExamDocumentTexts",
		Method:      http.MethodGet,
		Path:        "/patients/{patientId}/exam-document-texts",
		Summary:     "Listar textos extraídos de documentos de exame",
		Tags:        []string{"Exam documents"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
		Security:    security,
	}, h.listExamDocumentTexts)

	huma.Register(registered, huma.Operation{
		OperationID:   "uploadExamDocument",
		Method:        http.MethodPost,
		Path:          "/patients/{patientId}/exam-documents",
		Summary:       "Enviar documento de exame",
		Tags:          []string{"Exam documents"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType},
		Security:      security,
	}, h.uploadExamDocument)

	huma.Register(registered, huma.Operation{
		OperationID: "getExamDocument",
		Method:      http.MethodGet,
		Path:        "/exam-documents/{documentId}",
		Summary:     "Obter documento de exame",
		Tags:        []string{"Exam documents"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
		Security:    security,
	}, h.getExamDocument)

	huma.Register(registered, huma.Operation{
		OperationID: "getExamDocumentFile",
		Method:      http.MethodGet,
		Path:        "/exam-documents/{documentId}/file",
		Summary:     "Obter URL temporária do arquivo de exame",
		Tags:        []string{"Exam documents"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
		Security:    security,
	}, h.getExamDocumentFile)
}

func (h *ExamsHandler) listExamDocuments(ctx context.Context, input *listExamDocumentsInput) (*listExamDocumentsOutput, error) {
	currentUser, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.accessChecker.RequireAccess(ctx, currentUser.ID, input.PatientID); err != nil {
		return nil, toHumaError(err)
	}
	list, err := h.svc.ListByPatient(ctx, input.PatientID, input.Limit, input.Offset)
	if err != nil {
		return nil, toHumaError(err)
	}
	return &listExamDocumentsOutput{Body: list}, nil
}

func (h *ExamsHandler) listExamDocumentTexts(ctx context.Context, input *listExamDocumentsInput) (*listExamDocumentTextsOutput, error) {
	currentUser, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.accessChecker.RequireAccess(ctx, currentUser.ID, input.PatientID); err != nil {
		return nil, toHumaError(err)
	}
	list, err := h.svc.ListDocumentTextsByPatient(ctx, input.PatientID, input.Limit, input.Offset)
	if err != nil {
		return nil, toHumaError(err)
	}
	return &listExamDocumentTextsOutput{Body: list}, nil
}

func (h *ExamsHandler) getExamDocument(ctx context.Context, input *examDocumentInput) (*examDocumentOutput, error) {
	document, err := h.findAccessibleDocument(ctx, input.DocumentID)
	if err != nil {
		return nil, err
	}
	return &examDocumentOutput{Body: *document}, nil
}

func (h *ExamsHandler) getExamDocumentFile(ctx context.Context, input *examDocumentInput) (*examDocumentFileOutput, error) {
	document, err := h.findAccessibleDocument(ctx, input.DocumentID)
	if err != nil {
		return nil, err
	}
	if h.storage == nil {
		return nil, huma.Error500InternalServerError("armazenamento de documentos indisponível")
	}
	url, err := h.storage.GetSignedURL(ctx, document.StorageURI, examDocumentFileURLExpirationMinutes)
	if err != nil {
		return nil, toHumaError(err)
	}
	return &examDocumentFileOutput{Body: examDocumentFileResponse{
		URL:       url,
		ExpiresAt: time.Now().UTC().Add(examDocumentFileURLExpirationMinutes * time.Minute),
	}}, nil
}

func (h *ExamsHandler) uploadExamDocument(ctx context.Context, input *uploadExamDocumentInput) (*examDocumentOutput, error) {
	currentUser, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.accessChecker.RequireAccess(ctx, currentUser.ID, input.PatientID); err != nil {
		return nil, toHumaError(err)
	}
	collectionDate, err := parseExamCollectionDate(input.RawBody.Data().CollectionDate)
	if err != nil {
		return nil, toHumaError(err)
	}
	fileHeaders := input.RawBody.Form.File["file"]
	if len(fileHeaders) != 1 {
		return nil, huma.Error422UnprocessableEntity("arquivo é obrigatório")
	}
	upload, err := UploadDocument(ctx, fileHeaders[0], input.PatientID, h.storage)
	if err != nil {
		return nil, toHumaError(err)
	}
	defer os.Remove(upload.LocalPath)

	document, err := h.processor.Execute(ctx, processing.ProcessStoredDocumentInput{
		PatientID:        input.PatientID,
		UploadedByUserID: currentUser.ID,
		StorageURI:       upload.StorageURI,
		OriginalFilename: upload.OriginalFilename,
		MimeType:         upload.MimeType,
		LocalPath:        upload.LocalPath,
		CollectionDate:   collectionDate,
	})
	if err != nil {
		return nil, toHumaError(err)
	}
	return &examDocumentOutput{Body: *document}, nil
}

func (h *ExamsHandler) findAccessibleDocument(ctx context.Context, documentID uuid.UUID) (*documents.ExamDocumentOutput, error) {
	currentUser, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	document, err := h.svc.FindByID(ctx, documentID)
	if err != nil {
		return nil, toHumaError(err)
	}
	if document == nil {
		return nil, huma.Error404NotFound("documento não encontrado")
	}
	if err := h.accessChecker.RequireAccess(ctx, currentUser.ID, document.PatientID); err != nil {
		return nil, toHumaError(err)
	}
	return document, nil
}

func humaCurrentUser(ctx context.Context) (*accountdomain.User, error) {
	currentUser, ok := helpers.GetCurrentUserFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("conta registrada necessária")
	}
	return currentUser, nil
}

func toHumaError(err error) error {
	return humaerror.From(err)
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
