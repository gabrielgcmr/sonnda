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
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeExamsService struct {
	listByPatientCalled      bool
	listDocumentTextsCalled  bool
	createDocumentTextCalled bool
	createdInput             examsvc.CreateExamDocumentInput
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
	return nil, nil
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

func (f *fakeTextExtractor) Extract(ctx context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	f.input = input
	return &domaintext.ExtractOutput{
		Text:   "HEMOGRAMA\nHemoglobina 15,1 g/dL",
		Method: "test_text_extractor",
	}, nil
}

func TestDocumentTextForDisplayUsesNormalizedTextAndKeepsRawFallback(t *testing.T) {
	if got := documentTextForDisplay(&domaintext.ExtractOutput{Text: "Hematocrito 43,8 \uFF05", NormalizedText: "Hematocrito 43,8 %"}); got != "Hematocrito 43,8 %" {
		t.Fatalf("normalized display text = %q", got)
	}
	if got := documentTextForDisplay(&domaintext.ExtractOutput{Text: "Hematocrito 43,8 \uFF05"}); got != "Hematocrito 43,8 \uFF05" {
		t.Fatalf("raw fallback text = %q", got)
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
	h := NewExams(svc, nil, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:id/exames", h.ListExamDocuments)

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

func TestListExamDocumentTexts_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := NewExams(svc, nil, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:id/exames/document-texts", h.ListExamDocumentTexts)

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
	h := NewExams(svc, nil, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:id/exames", h.ListExamDocuments)

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
	h := NewExams(svc, createLabUC, storage, textExtractor, allowAllAuthorizer{})

	body, contentType := multipartBody(t, "file", "hemograma.pdf", "application/pdf", []byte("%PDF-1.7\nfake"))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: userID, AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.POST("/v1/patients/:id/exames", h.UploadExamDocument)

	req := httptest.NewRequest(http.MethodPost, "/v1/patients/"+patientID.String()+"/exames", body)
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
	if !svc.createDocumentTextCalled {
		t.Fatal("expected structured lab output to also create exam document text")
	}
	if svc.documentTextInput.Category != exams.ExamTypeLaboratory {
		t.Fatalf("expected document text category laboratory, got %q", svc.documentTextInput.Category)
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
			h := NewExams(svc, &fakeCreateLabReportUC{err: tc.err}, &fakeExamStorage{}, &fakeTextExtractor{}, allowAllAuthorizer{})
			r := gin.New()
			r.Use(func(c *gin.Context) {
				helpers.SetCurrentUser(c, &user.User{ID: uuid.New(), AccountType: user.AccountTypeBasicCare})
				c.Next()
			})
			r.POST("/v1/patients/:id/exames", h.UploadExamDocument)
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

func TestBuildLabText_FormatsStructuredResults(t *testing.T) {
	reportDate := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	patientName := "Gabriel Cactus Moreno Reboucas"
	labName := "Laboratorio Exemplo"
	unit := "g/dL"
	value := "15,1"
	reference := "13,5 a 17,5"

	text := buildLabReportText(&labsvc.LabReportOutput{
		PatientName: &patientName,
		LabName:     &labName,
		ReportDate:  &reportDate,
		TestResults: []labsvc.TestResultOutput{
			{
				TestName: "HEMOGRAMA",
				Items: []labsvc.TestItemOutput{
					{
						ParameterName: "Hemoglobina",
						ResultValue:   &value,
						ResultUnit:    &unit,
						ReferenceText: &reference,
					},
				},
			},
		},
	})

	expectedParts := []string{
		"Exame laboratorial",
		"Paciente: Gabriel Cactus Moreno Reboucas",
		"Laboratorio: Laboratorio Exemplo",
		"Data do laudo: 31/08/2026",
		"HEMOGRAMA",
		"- Hemoglobina 15,1 g/dL (Referencia: 13,5 a 17,5)",
	}
	for _, part := range expectedParts {
		if !strings.Contains(text, part) {
			t.Fatalf("expected text to contain %q, got:\n%s", part, text)
		}
	}
}

func multipartBody(t *testing.T, fieldName, filename, contentType string, contents []byte) (*bytes.Buffer, string) {
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
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}
