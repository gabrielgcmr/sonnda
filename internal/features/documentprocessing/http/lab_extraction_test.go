// internal/features/documentprocessing/http/lab_extraction_test.go
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/extraction"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"
	domaintext "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/textextraction"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type temporaryTextExtractorStub struct{ path string }

func (s *temporaryTextExtractorStub) Extract(_ context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	s.path = input.LocalPath
	if _, err := os.Stat(input.LocalPath); err != nil {
		return nil, err
	}
	return &domaintext.ExtractOutput{Text: "Glicose 90 mg/dL", Method: "pdf_text_raw"}, nil
}

type temporaryLabExtractorStub struct {
	called bool
	err    error
}

func (s *temporaryLabExtractorStub) ExtractLabReport(_ context.Context, input labextraction.ExtractLabReportInput) (*labextraction.ExtractedLabReport, error) {
	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	value, unit := "90", "mg/dL"
	return &labextraction.ExtractedLabReport{Tests: []labextraction.ExtractedTestResult{{
		TestName: "Glicose",
		Items:    []labextraction.ExtractedTestItem{{ParameterName: "Glicose", ResultValue: &value, ResultUnit: &unit}},
	}}}, nil
}

func TestStandaloneLabExtractionDiscardsPDFAndReturnsStructuredResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	textExtractor := &temporaryTextExtractorStub{}
	labExtractor := &temporaryLabExtractorStub{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(helpers.ContextWithCurrentUser(c.Request.Context(), &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare}))
		c.Next()
	})
	api := humagin.New(router, huma.DefaultConfig("test", "test"))
	NewStandaloneLabExtraction(extraction.New(textExtractor, labExtractor)).RegisterHumaRoutes(api, nil)

	body, contentType := temporaryLabMultipart(t, "exam.pdf", "application/pdf", []byte("%PDF-1.4"))
	request := httptest.NewRequest(http.MethodPost, "/lab-extractions", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if !labExtractor.called || textExtractor.path == "" {
		t.Fatal("temporary extraction did not call both extractors")
	}
	if _, err := os.Stat(textExtractor.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary PDF was not removed: %v", err)
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("Glicose")) {
		t.Fatalf("missing structured response: %s", response.Body.String())
	}
}

func TestStandaloneLabExtractionRejectsNonPDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(helpers.ContextWithCurrentUser(c.Request.Context(), &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare}))
		c.Next()
	})
	api := humagin.New(router, huma.DefaultConfig("test", "test"))
	NewStandaloneLabExtraction(extraction.New(&temporaryTextExtractorStub{}, &temporaryLabExtractorStub{})).RegisterHumaRoutes(api, nil)
	body, contentType := temporaryLabMultipart(t, "exam.txt", "text/plain", []byte("not a PDF"))
	request := httptest.NewRequest(http.MethodPost, "/lab-extractions", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
}

func TestTemporaryExtractionFailureStillRemovesPDF(t *testing.T) {
	reader := &temporaryTextExtractorStub{}
	provider := &temporaryLabExtractorStub{err: errors.New("provider unavailable")}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(helpers.ContextWithCurrentUser(c.Request.Context(), &accountdomain.User{ID: uuid.New()}))
	})
	NewStandaloneLabExtraction(extraction.New(reader, provider)).RegisterHumaRoutes(humagin.New(router, huma.DefaultConfig("test", "test")), nil)
	body, contentType := temporaryLabMultipart(t, "exam.pdf", "application/pdf", []byte("%PDF-1.4"))
	request := httptest.NewRequest(http.MethodPost, "/lab-extractions", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code < 500 || bytes.Contains(response.Body.Bytes(), []byte("provider unavailable")) {
		t.Fatalf("unsafe failure response: %d %s", response.Code, response.Body.String())
	}
	if _, err := os.Stat(reader.path); !os.IsNotExist(err) {
		t.Fatal("temporary PDF leaked on failure")
	}
}

func TestTemporaryPDFValidation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content []byte
		status  int
	}{
		{"empty", nil, 400},
		{"renamed text", []byte("not a PDF"), 400},
		{"over limit", append([]byte("%PDF-"), make([]byte, temporaryLabExtractionMaxFileSize)...), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Request = c.Request.WithContext(helpers.ContextWithCurrentUser(c.Request.Context(), &accountdomain.User{ID: uuid.New()}))
			})
			NewStandaloneLabExtraction(extraction.New(&temporaryTextExtractorStub{}, &temporaryLabExtractorStub{})).RegisterHumaRoutes(humagin.New(router, huma.DefaultConfig("test", "test")), nil)
			body, contentType := temporaryLabMultipart(t, "exam.pdf", "application/pdf", tc.content)
			request := httptest.NewRequest(http.MethodPost, "/lab-extractions", body)
			request.Header.Set("Content-Type", contentType)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

type pdfExtractorFunc func(context.Context, string, string) (*extraction.Result, error)

func (f pdfExtractorFunc) ExtractPDF(ctx context.Context, path, filename string) (*extraction.Result, error) {
	return f(ctx, path, filename)
}

func TestStandaloneLabExtractionUsesInjectedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	want := extraction.Result{
		Status:      labextraction.ExtractionStatusPartial,
		Warnings:    []labextraction.ExtractionWarning{{Code: "review", Message: "Conferir resultado."}},
		SummaryText: "Resumo retornado pelo serviço injetado.",
	}
	user := &accountdomain.User{ID: uuid.New()}
	requestContext := helpers.ContextWithCurrentUser(context.Background(), user)
	var temporaryPath string
	calls := 0
	extractor := pdfExtractorFunc(func(ctx context.Context, path, filename string) (*extraction.Result, error) {
		calls++
		temporaryPath = path
		if ctx != requestContext || filename != "injected.pdf" {
			t.Fatalf("request context or filename changed: %s", filename)
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "%PDF-1.4" {
			t.Fatalf("temporary file unavailable to injected service: %q, %v", data, err)
		}
		return &want, nil
	})
	router := gin.New()
	NewStandaloneLabExtraction(extractor).RegisterHumaRoutes(humagin.New(router, huma.DefaultConfig("test", "test")), nil)
	body, contentType := temporaryLabMultipart(t, "injected.pdf", "application/pdf", []byte("%PDF-1.4"))
	request := httptest.NewRequest(http.MethodPost, "/lab-extractions", body).WithContext(requestContext)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || calls != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code, calls, response.Body.String())
	}
	var got extraction.Result
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("injected result changed: got %+v, want %+v", got, want)
	}
	if _, err := os.Stat(temporaryPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary PDF was not removed: %v", err)
	}
}

func temporaryLabMultipart(t *testing.T, filename, _ string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}
