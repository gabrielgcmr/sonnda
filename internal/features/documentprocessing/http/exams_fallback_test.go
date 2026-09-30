// internal/features/documentprocessing/http/exams_fallback_test.go
package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	textsvc "github.com/gabrielgcmr/sonnda/internal/application/services/textextraction"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type recoveredTextExtractor struct{ input domaintext.ExtractInput }

func (f *recoveredTextExtractor) Extract(_ context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	f.input = input
	return &domaintext.ExtractOutput{Text: "GLICOSE Resultado: 90 mg/dL. Valor de referencia: 70 a 99 mg/dL. Material: sangue. Metodo: Hexoquinase.", Method: "document_ai_ocr"}, nil
}

func TestUploadContinuesAfterCloudOCRRecoversLocalFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeExamsService{}
	lab := &fakeCreateLabReportUC{}
	cloud := &recoveredTextExtractor{}
	storage := &fakeExamStorage{uri: "gs://bucket/photo.jpg"}
	extractor := textsvc.NewFallbackExtractor(reviewTextExtractor{errors.New("local OCR failed")}, cloud)
	h := newExamsHandler(svc, lab, storage, extractor, allowAllAccessChecker{})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(helpers.ContextWithCurrentUser(c.Request.Context(), &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare}))
		c.Next()
	})
	h.RegisterHumaRoutes(humagin.New(r, huma.DefaultConfig("test", "test")), nil)
	body, contentType := multipartBody(t, "file", "photo.jpg", "image/jpeg", []byte{0xff, 0xd8, 0xff, 0xe0})
	req := httptest.NewRequest(http.MethodPost, "/patients/"+uuid.NewString()+"/exam-documents", body)
	req.Header.Set("Content-Type", contentType)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("unexpected status %d: %s", resp.Code, resp.Body.String())
	}
	if cloud.input.DocumentURI != storage.uri || cloud.input.MimeType != "image/jpeg" {
		t.Fatalf("incorrect cloud input: %+v", cloud.input)
	}
	if svc.routeInput.ExtractionMethod != "document_ai_ocr" || svc.routeInput.ProcessingError != nil || svc.routeInput.ExtractedText == "" {
		t.Fatal("cloud text was not routed")
	}
	if !lab.called || !svc.createDocumentTextCalled {
		t.Fatal("recovered upload did not continue to laboratory processing")
	}
}
