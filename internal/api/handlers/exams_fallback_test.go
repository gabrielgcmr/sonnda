// internal/api/handlers/exams_fallback_test.go
package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	textsvc "github.com/gabrielgcmr/sonnda/internal/application/services/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
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
	h := NewExams(svc, lab, storage, extractor, allowAllAuthorizer{})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.New(), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.POST("/v1/patients/:id/exames", h.UploadExamDocument)
	body, contentType := multipartBody(t, "file", "photo.jpg", "image/jpeg", []byte{0xff, 0xd8, 0xff, 0xe0})
	req := httptest.NewRequest(http.MethodPost, "/v1/patients/"+uuid.NewString()+"/exames", body)
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
