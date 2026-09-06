// internal/application/services/exams/review_test.go
package examsvc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type reviewRepository struct {
	repository.Exams
	document exams.ExamDocument
}

func (r *reviewRepository) FindByID(context.Context, uuid.UUID) (*exams.ExamDocument, error) {
	return &r.document, nil
}

func (r *reviewRepository) MarkClassified(_ context.Context, _ uuid.UUID, status exams.DocumentStatus, examType exams.ExamType, method *string, confidence *float64, text, message *string) (*exams.ExamDocument, error) {
	r.document.Status = status
	r.document.ExamType = &examType
	r.document.ExtractionMethod = method
	r.document.Confidence = confidence
	r.document.ExtractedText = text
	r.document.ErrorMessage = message
	return &r.document, nil
}

func TestRouteDocumentPersistsReviewReasonAndClearsItAfterSuccess(t *testing.T) {
	for _, tc := range []struct {
		name    string
		text    string
		failure *apperr.AppError
		want    string
	}{
		{"unreadable", "", nil, "ler o texto"},
		{"unknown", "documento sem sinais de exame", nil, "identificar o tipo"},
		{"timeout", "", apperr.Internal("Tempo limite excedido.", errors.New("private detail")), "Tempo limite"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &reviewRepository{document: exams.ExamDocument{ID: uuid.New(), MimeType: "image/jpeg", OriginalFilename: "foto.jpg"}}
			svc := New(nil, repo)
			out, err := svc.RouteDocument(context.Background(), RouteExamDocumentInput{ID: repo.document.ID, ExtractedText: tc.text, ExtractionMethod: "metadata", ProcessingError: tc.failure})
			if err != nil {
				t.Fatal(err)
			}
			if out.Status != exams.DocumentStatusNeedsReview || out.ErrorMessage == nil || !strings.Contains(*out.ErrorMessage, tc.want) {
				t.Fatalf("missing review reason: %+v", out)
			}
			if strings.Contains(*out.ErrorMessage, "private detail") {
				t.Fatal("internal cause exposed")
			}
			out, err = svc.RouteDocument(context.Background(), RouteExamDocumentInput{ID: repo.document.ID, ExtractedText: "GLICOSE RESULTADO 90 mg/dL Valor de referencia 70 a 99 mg/dL Material: sangue", ExtractionMethod: "ocr_psm_6"})
			if err != nil {
				t.Fatal(err)
			}
			if out.Status != exams.DocumentStatusProcessed || repo.document.ErrorMessage != nil || out.ErrorMessage != nil {
				t.Fatalf("successful retry retained review: %+v", out)
			}
		})
	}
}
