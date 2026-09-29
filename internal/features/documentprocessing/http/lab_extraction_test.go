// internal/features/documentprocessing/http/lab_extraction_test.go
package http

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
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

func TestTemporaryLabExtractionDiscardsPDFAndReturnsStructuredResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	textExtractor := &temporaryTextExtractorStub{}
	labExtractor := &temporaryLabExtractorStub{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	api := humagin.New(router, huma.DefaultConfig("test", "test"))
	NewTemporaryLabExtraction(textExtractor, labExtractor).RegisterHumaRoutes(api, nil)

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

func TestTemporaryLabExtractionRejectsNonPDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	api := humagin.New(router, huma.DefaultConfig("test", "test"))
	NewTemporaryLabExtraction(&temporaryTextExtractorStub{}, &temporaryLabExtractorStub{}).RegisterHumaRoutes(api, nil)
	body, contentType := temporaryLabMultipart(t, "exam.txt", "text/plain", []byte("not a PDF"))
	request := httptest.NewRequest(http.MethodPost, "/lab-extractions", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
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
