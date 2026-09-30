// internal/features/documentprocessing/http/review.go
package http

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/extraction"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	"net/http"
)

type extractionOutput struct{ Body extraction.Result }
type confirmationOutput struct{ Body laboratory.LabReportOutput }

func (h *ExamsHandler) registerReviewRoutes(api huma.API, security []map[string][]string) {
	huma.Register(api, huma.Operation{OperationID: "getExamDocumentExtraction", Method: http.MethodGet, Path: "/exam-documents/{documentId}/extraction", Summary: "Conferir extração do documento", Tags: []string{"Exam documents"}, Security: security, Errors: []int{401, 403, 404}}, h.getExtraction)
	huma.Register(api, huma.Operation{OperationID: "confirmExamDocument", Method: http.MethodPost, Path: "/exam-documents/{documentId}/confirmation", Summary: "Confirmar exame no histórico", Tags: []string{"Exam documents"}, Security: security, Errors: []int{401, 403, 404, 409, 422}}, h.confirm)
	huma.Register(api, huma.Operation{OperationID: "discardExamDocument", Method: http.MethodDelete, Path: "/exam-documents/{documentId}", Summary: "Excluir rascunho e PDF", DefaultStatus: http.StatusNoContent, Tags: []string{"Exam documents"}, Security: security, Errors: []int{401, 403, 404, 409}}, h.discard)
}

func (h *ExamsHandler) getExtraction(ctx context.Context, input *examDocumentInput) (*extractionOutput, error) {
	if _, err := h.findAccessibleDocument(ctx, input.DocumentID); err != nil {
		return nil, err
	}
	result, err := h.drafts.Extraction(ctx, input.DocumentID)
	if err != nil {
		return nil, toHumaError(ctx, err)
	}
	return &extractionOutput{Body: *result}, nil
}

func (h *ExamsHandler) confirm(ctx context.Context, input *examDocumentInput) (*confirmationOutput, error) {
	if _, err := h.findAccessibleDocument(ctx, input.DocumentID); err != nil {
		return nil, err
	}
	user, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	report, err := h.confirmer.Confirm(ctx, input.DocumentID, user.ID)
	if err != nil {
		return nil, toHumaError(ctx, err)
	}
	return &confirmationOutput{Body: *report}, nil
}

func (h *ExamsHandler) discard(ctx context.Context, input *examDocumentInput) (*struct{}, error) {
	// A missing resource remains 404, allowing clients to treat repeated deletion as complete.
	if _, err := h.findAccessibleDocument(ctx, input.DocumentID); err != nil {
		return nil, err
	}
	if err := h.drafts.Delete(ctx, input.DocumentID); err != nil {
		return nil, toHumaError(ctx, err)
	}
	return nil, nil
}
