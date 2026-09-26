// internal/api/handlers/exams_test.go
package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	examsvc "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	exams "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	labsuc "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	labsvc "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeExamsService struct {
	listByPatientCalled      bool
	listDocumentTextsCalled  bool
	createDocumentTextCalled bool
	createdInput             examsvc.CreateExamDocumentInput
	document                 *examsvc.ExamDocumentOutput
	routeInput               examsvc.RouteExamDocumentInput
	documentTextInput        examsvc.CreateExamDocumentTextFromTextInput
	failedInput              examsvc.MarkExamDocumentFailedInput
	patientID                uuid.UUID
	limit                    int
	offset                   int
}

func (f *fakeExamsService) Create(ctx context.Context, input examsvc.CreateExamDocumentInput) (*examsvc.ExamDocumentOutput, error) {
	f.createdInput = input
	now := time.Now().UTC()
	return &examsvc.ExamDocumentOutput{
		ID:               uuid.Must(uuid.NewV7()),
		PatientID:        input.PatientID,
		UploadedByUserID: input.UploadedByUserID,
		StorageURI:       input.StorageURI,
		OriginalFilename: input.OriginalFilename,
		MimeType:         input.MimeType,
		Status:           exams.DocumentStatusUploaded,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (f *fakeExamsService) FindByID(ctx context.Context, id uuid.UUID) (*examsvc.ExamDocumentOutput, error) {
	return f.document, nil
}

func (f *fakeExamsService) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]examsvc.ExamDocumentOutput, error) {
	f.listByPatientCalled = true
	f.patientID = patientID
	f.limit = limit
	f.offset = offset
	return []examsvc.ExamDocumentOutput{}, nil
}

func (f *fakeExamsService) RouteDocument(ctx context.Context, input examsvc.RouteExamDocumentInput) (*examsvc.ExamDocumentOutput, error) {
	f.routeInput = input
	now := time.Now().UTC()
	examType := exams.ExamTypeLaboratory
	method := input.ExtractionMethod
	confidence := 0.90
	status := exams.DocumentStatusProcessed
	var reviewMessage *string
	if input.ProcessingError != nil {
		status = exams.DocumentStatusNeedsReview
		message := input.ProcessingError.Message
		reviewMessage = &message
	}
	return &examsvc.ExamDocumentOutput{
		ID:               input.ID,
		PatientID:        f.createdInput.PatientID,
		UploadedByUserID: f.createdInput.UploadedByUserID,
		StorageURI:       f.createdInput.StorageURI,
		OriginalFilename: f.createdInput.OriginalFilename,
		MimeType:         f.createdInput.MimeType,
		Status:           status,
		ErrorMessage:     reviewMessage,
		ExamType:         &examType,
		ExtractionMethod: &method,
		Confidence:       &confidence,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (f *fakeExamsService) MarkFailed(ctx context.Context, input examsvc.MarkExamDocumentFailedInput) (*examsvc.ExamDocumentOutput, error) {
	f.failedInput = input
	return nil, nil
}

func (f *fakeExamsService) CreateDocumentTextFromText(ctx context.Context, input examsvc.CreateExamDocumentTextFromTextInput) (*examsvc.ExamDocumentTextOutput, error) {
	f.createDocumentTextCalled = true
	f.documentTextInput = input
	now := time.Now().UTC()
	return &examsvc.ExamDocumentTextOutput{
		ID:               uuid.Must(uuid.NewV7()),
		ExamDocumentID:   &input.ExamDocumentID,
		PatientID:        input.PatientID,
		UploadedByUserID: input.UploadedByUserID,
		Category:         input.Category,
		Text:             input.Text,
		PerformedAt:      input.PerformedAt,
		ExtractionMethod: &input.ExtractionMethod,
		Confidence:       input.Confidence,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (f *fakeExamsService) ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]examsvc.ExamDocumentTextOutput, error) {
	f.listDocumentTextsCalled = true
	f.patientID = patientID
	f.limit = limit
	f.offset = offset
	return []examsvc.ExamDocumentTextOutput{}, nil
}

type fakeExamStorage struct {
	uri         string
	objectName  string
	contentType string
}

func (f *fakeExamStorage) Upload(ctx context.Context, file io.Reader, objectName, contentType string) (string, error) {
	f.objectName = objectName
	f.contentType = contentType
	if f.uri != "" {
		return f.uri, nil
	}
	return "gs://bucket/exam.pdf", nil
}

func (f *fakeExamStorage) Delete(ctx context.Context, uri string) error {
	return nil
}

func (f *fakeExamStorage) GetSignedURL(ctx context.Context, uri string, expirationMinutes int) (string, error) {
	return "", nil
}

type fakeTextExtractor struct {
	input domaintext.ExtractInput
}

type testLaboratoryProcessor struct {
	creator   labsuc.CreateLabReportFromDocumentUseCase
	documents interface {
		CreateDocumentTextFromText(context.Context, examsvc.CreateExamDocumentTextFromTextInput) (*examsvc.ExamDocumentTextOutput, error)
	}
}

func (testLaboratoryProcessor) Kind() labsuc.DocumentKind { return labsuc.DocumentKindLaboratory }

func (p testLaboratoryProcessor) Process(ctx context.Context, input labsuc.ProcessingInput) error {
	report, err := p.creator.Execute(ctx, labsuc.CreateLabReportFromDocumentInput{
		PatientID:        input.Document.PatientID,
		ExamDocumentID:   &input.Document.ID,
		DocumentURI:      input.Document.StorageURI,
		MimeType:         input.Document.MimeType,
		UploadedByUserID: input.Document.UploadedByUserID,
		CollectionDate:   input.CollectionDate,
	})
	if err != nil || report == nil {
		return err
	}
	var summary strings.Builder
	for _, result := range report.TestResults {
		for _, item := range result.Items {
			if item.ResultValue != nil && item.ResultUnit != nil {
				summary.WriteString("- " + item.ParameterName + " " + *item.ResultValue + " " + *item.ResultUnit + "\n")
			}
		}
	}
	_, err = p.documents.CreateDocumentTextFromText(ctx, examsvc.CreateExamDocumentTextFromTextInput{
		ExamDocumentID:   input.Document.ID,
		PatientID:        input.Document.PatientID,
		UploadedByUserID: input.Document.UploadedByUserID,
		Category:         exams.ExamTypeLaboratory,
		Text:             strings.TrimSpace(summary.String()),
		PerformedAt:      input.CollectionDate,
		ExtractionMethod: "test_laboratory_processor",
	})
	return err
}

func newExamsHandler(
	svc examsvc.Service,
	laboratory labsuc.CreateLabReportFromDocumentUseCase,
	storage domainstorage.FileStorageService,
	extractor domaintext.Extractor,
	accessChecker patientaccess.Checker,
) *ExamsHandler {
	var processors []labsuc.Processor
	if laboratory != nil {
		processors = append(processors, testLaboratoryProcessor{creator: laboratory, documents: svc})
	}
	return NewExams(svc, labsuc.NewProcessStoredDocument(svc, extractor, processors...), storage, accessChecker)
}

func (f *fakeTextExtractor) Extract(ctx context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	f.input = input
	return &domaintext.ExtractOutput{
		Text:   "HEMOGRAMA\nHemoglobina 15,1 g/dL",
		Method: "test_text_extractor",
	}, nil
}

func TestParseExamCollectionDate(t *testing.T) {
	date, err := parseExamCollectionDate("2026-09-16")
	if err != nil || date == nil {
		t.Fatalf("valid date returned date=%v err=%v", date, err)
	}
	want := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	if !date.Equal(want) {
		t.Fatalf("date = %v, want %v", date, want)
	}

	empty, err := parseExamCollectionDate(" ")
	if err != nil || empty != nil {
		t.Fatalf("optional empty date returned date=%v err=%v", empty, err)
	}

	if _, err := parseExamCollectionDate("16/09/2026"); err == nil {
		t.Fatal("expected invalid API date format to fail")
	}
}

type fakeCreateLabReportUC struct {
	called bool
	input  labsuc.CreateLabReportFromDocumentInput
	err    error
}

func (f *fakeCreateLabReportUC) Execute(ctx context.Context, input labsuc.CreateLabReportFromDocumentInput) (*labsvc.LabReportOutput, error) {
	f.called = true
	f.input = input
	if f.err != nil {
		return nil, f.err
	}

	value := "15,1"
	unit := "g/dL"
	return &labsvc.LabReportOutput{
		ID:               uuid.Must(uuid.NewV7()),
		PatientID:        input.PatientID,
		ExamDocumentID:   input.ExamDocumentID,
		UploadedByUserID: input.UploadedByUserID,
		TestResults: []labsvc.TestResultOutput{
			{
				ID:       uuid.Must(uuid.NewV7()),
				TestName: "HEMOGRAMA",
				Items: []labsvc.TestItemOutput{
					{
						ID:            uuid.Must(uuid.NewV7()),
						ParameterName: "Hemoglobina",
						ResultValue:   &value,
						ResultUnit:    &unit,
					},
				},
			},
		},
	}, nil
}

func TestListExamDocuments_UsesServiceWithDefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := newExamsHandler(svc, nil, nil, nil, allowAllAccessChecker{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.Must(uuid.NewV7()), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:patientId/exames", h.ListExamDocuments)

	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id.String()+"/exames", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listByPatientCalled {
		t.Fatal("expected ListByPatient to be called")
	}
	if svc.patientID != id {
		t.Fatalf("expected patientID %s, got %s", id, svc.patientID)
	}
	if svc.limit != 100 {
		t.Fatalf("expected default limit 100, got %d", svc.limit)
	}
	if svc.offset != 0 {
		t.Fatalf("expected default offset 0, got %d", svc.offset)
	}
}

func TestGetExamDocumentReturnsDocumentByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	documentID := uuid.New()
	document := &examsvc.ExamDocumentOutput{ID: documentID, PatientID: uuid.New()}
	svc := &fakeExamsService{document: document}
	access := &recordingExamAccessChecker{}
	h := newExamsHandler(svc, nil, nil, nil, access)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/exam-documents/:documentId", h.GetExamDocument)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/exam-documents/"+documentID.String(), nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), documentID.String()) {
		t.Fatalf("response does not contain document id %s: %s", documentID, response.Body.String())
	}
	if access.patientID != document.PatientID {
		t.Fatalf("access checked for patient %s, want %s", access.patientID, document.PatientID)
	}
}

type recordingExamAccessChecker struct{ patientID uuid.UUID }

func (a *recordingExamAccessChecker) RequireAccess(_ context.Context, _, patientID uuid.UUID) error {
	a.patientID = patientID
	return nil
}

func TestListExamDocumentTexts_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := newExamsHandler(svc, nil, nil, nil, allowAllAccessChecker{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.Must(uuid.NewV7()), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:patientId/exames/document-texts", h.ListExamDocumentTexts)

	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id.String()+"/exames/document-texts", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listDocumentTextsCalled {
		t.Fatal("expected ListDocumentTextsByPatient to be called")
	}
}

func TestListExamDocuments_UsesQueryPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := newExamsHandler(svc, nil, nil, nil, allowAllAccessChecker{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.Must(uuid.NewV7()), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:patientId/exames", h.ListExamDocuments)

	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id.String()+"/exames?limit=25&offset=50", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if svc.limit != 25 {
		t.Fatalf("expected limit 25, got %d", svc.limit)
	}
	if svc.offset != 50 {
		t.Fatalf("expected offset 50, got %d", svc.offset)
	}
}

func TestUploadExamDocument_WhenClassifiedAsLab_UsesStructuredLabPipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())
	storage := &fakeExamStorage{uri: "gs://bucket/patients/exam.pdf"}
	textExtractor := &fakeTextExtractor{}
	createLabUC := &fakeCreateLabReportUC{}
	svc := &fakeExamsService{}
	h := newExamsHandler(svc, createLabUC, storage, textExtractor, allowAllAccessChecker{})

	body, contentType := multipartBodyWithFields(
		t,
		"file",
		"hemograma.pdf",
		"application/pdf",
		[]byte("%PDF-1.7\nfake"),
		map[string]string{"collection_date": "2026-09-16"},
	)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: userID, AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	r.POST("/v1/patients/:patientId/exam-documents", h.UploadExamDocument)

	req := httptest.NewRequest(http.MethodPost, "/v1/patients/"+patientID.String()+"/exam-documents", body)
	req.Header.Set("Content-Type", contentType)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
	}
	if !createLabUC.called {
		t.Fatal("expected laboratory upload to use CreateLabReportFromDocumentUseCase")
	}
	if createLabUC.input.PatientID != patientID {
		t.Fatalf("expected patientID %s, got %s", patientID, createLabUC.input.PatientID)
	}
	if createLabUC.input.UploadedByUserID != userID {
		t.Fatalf("expected uploadedByUserID %s, got %s", userID, createLabUC.input.UploadedByUserID)
	}
	if createLabUC.input.ExamDocumentID == nil || *createLabUC.input.ExamDocumentID != svc.routeInput.ID {
		t.Fatalf("expected lab report to be linked to exam document %s, got %v", svc.routeInput.ID, createLabUC.input.ExamDocumentID)
	}
	if createLabUC.input.DocumentURI != storage.uri {
		t.Fatalf("expected document URI %q, got %q", storage.uri, createLabUC.input.DocumentURI)
	}
	if createLabUC.input.MimeType != "application/pdf" {
		t.Fatalf("expected mime type application/pdf, got %q", createLabUC.input.MimeType)
	}
	wantCollectionDate := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	if createLabUC.input.CollectionDate == nil || !createLabUC.input.CollectionDate.Equal(wantCollectionDate) {
		t.Fatalf("expected collection date %v, got %v", wantCollectionDate, createLabUC.input.CollectionDate)
	}
	if !svc.createDocumentTextCalled {
		t.Fatal("expected structured lab output to also create exam document text")
	}
	if svc.documentTextInput.Category != exams.ExamTypeLaboratory {
		t.Fatalf("expected document text category laboratory, got %q", svc.documentTextInput.Category)
	}
	if svc.documentTextInput.PerformedAt == nil || !svc.documentTextInput.PerformedAt.Equal(wantCollectionDate) {
		t.Fatalf("expected document text date %v, got %v", wantCollectionDate, svc.documentTextInput.PerformedAt)
	}
	if !strings.Contains(svc.documentTextInput.Text, "- Hemoglobina 15,1 g/dL") {
		t.Fatalf("expected document text to include structured lab item, got:\n%s", svc.documentTextInput.Text)
	}
}

func TestUploadExamDocument_LabFailureIsNotReportedAsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"duplicate", apperr.AlreadyExists("exame ja cadastrado"), http.StatusConflict},
		{"database", apperr.Internal("falha ao salvar exame", errors.New("private database detail")), http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeExamsService{}
			h := newExamsHandler(svc, &fakeCreateLabReportUC{err: tc.err}, &fakeExamStorage{}, &fakeTextExtractor{}, allowAllAccessChecker{})
			r := gin.New()
			r.Use(func(c *gin.Context) {
				helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare})
				c.Next()
			})
			r.POST("/v1/patients/:patientId/exames", h.UploadExamDocument)
			body, contentType := multipartBody(t, "file", "exam.pdf", "application/pdf", []byte("%PDF-1.7\ntest"))
			req := httptest.NewRequest(http.MethodPost, "/v1/patients/"+uuid.NewString()+"/exames", body)
			req.Header.Set("Content-Type", contentType)
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)
			if resp.Code != tc.status {
				t.Fatalf("expected %d, got %d: %s", tc.status, resp.Code, resp.Body.String())
			}
			if svc.failedInput.ID != svc.routeInput.ID || svc.failedInput.ID == uuid.Nil {
				t.Fatal("uploaded document was not marked failed")
			}
			if svc.createDocumentTextCalled {
				t.Fatal("text must not be created from a failed lab extraction")
			}
			if strings.Contains(resp.Body.String(), "private database detail") || strings.Contains(svc.failedInput.ErrorMessage, "private database detail") {
				t.Fatal("internal error details were exposed")
			}
		})
	}
}

func multipartBody(t *testing.T, fieldName, filename, contentType string, contents []byte) (*bytes.Buffer, string) {
	t.Helper()
	return multipartBodyWithFields(t, fieldName, filename, contentType, contents, nil)
}

func multipartBodyWithFields(
	t *testing.T,
	fieldName, filename, contentType string,
	contents []byte,
	fields map[string]string,
) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename)},
		"Content-Type":        {contentType},
	})
	if err != nil {
		t.Fatalf("failed to create multipart part: %v", err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatalf("failed to write multipart body: %v", err)
	}
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatalf("failed to write multipart field %s: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}
