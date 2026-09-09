// internal/api/handlers/exams_review_test.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type reviewTextExtractor struct{ err error }

func (e reviewTextExtractor) Extract(context.Context, domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	return nil, e.err
}

func TestUploadExamDocumentExplainsOCRReview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name      string
		extractor domaintext.Extractor
		want      string
	}{
		{"timeout", reviewTextExtractor{fmt.Errorf("ocr: %w", context.DeadlineExceeded)}, "tempo limite"},
		{"failure", reviewTextExtractor{errors.New("private OCR detail")}, "ler o texto"},
		{"nil output", reviewTextExtractor{}, "texto legivel"},
		{"unavailable", nil, "indisponivel"},
		{"all attempts failed", reviewTextExtractor{apperr.Internal("Falha apos duas tentativas. O arquivo precisa de revisao.", errors.New("private OCR detail"))}, "duas tentativas"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeExamsService{}
			lab := &fakeCreateLabReportUC{}
			h := NewExams(svc, lab, &fakeExamStorage{}, tc.extractor, allowAllAuthorizer{})
			r := gin.New()
			r.Use(func(c *gin.Context) {
				helpers.SetCurrentUser(c, &user.User{ID: uuid.New(), AccountType: user.AccountTypeBasicCare})
				c.Next()
			})
			r.POST("/v1/patients/:id/exames", h.UploadExamDocument)
			body, contentType := multipartBody(t, "file", "hemograma.jpg", "image/jpeg", []byte{0xff, 0xd8, 0xff, 0xe0})
			req := httptest.NewRequest(http.MethodPost, "/v1/patients/"+uuid.NewString()+"/exames", body)
			req.Header.Set("Content-Type", contentType)
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)
			if resp.Code != http.StatusCreated {
				t.Fatalf("status %d: %s", resp.Code, resp.Body.String())
			}
			var out examsvc.ExamDocumentOutput
			if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Status != exams.DocumentStatusNeedsReview || out.ErrorMessage == nil || !strings.Contains(*out.ErrorMessage, tc.want) {
				t.Fatalf("missing review reason: %s", resp.Body.String())
			}
			if strings.Contains(resp.Body.String(), "private OCR detail") {
				t.Fatal("internal cause exposed")
			}
			if svc.routeInput.ExtractionMethod != "metadata" || lab.called {
				t.Fatal("unexpected processing after OCR failure")
			}
		})
	}
}
