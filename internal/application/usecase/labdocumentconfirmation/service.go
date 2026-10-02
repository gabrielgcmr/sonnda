// internal/application/usecase/labdocumentconfirmation/service.go
package labdocumentconfirmation

import (
	"context"
	"errors"

	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	domain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/extraction"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type Repository interface {
	Confirm(context.Context, uuid.UUID, uuid.UUID, *labs.LabReport, string, *string) (uuid.UUID, error)
}
type Service struct {
	documents  documents.Service
	drafts     *documents.Drafts
	repository Repository
	labs       laboratory.Repository
}

func New(documents documents.Service, drafts *documents.Drafts, repo Repository, labs laboratory.Repository) *Service {
	return &Service{documents: documents, drafts: drafts, repository: repo, labs: labs}
}

func (s *Service) Confirm(ctx context.Context, id, userID uuid.UUID) (*laboratory.LabReportOutput, error) {
	document, err := s.documents.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, apperr.NotFound("Documento não encontrado.")
	}
	if document.ReviewStatus == nil || (*document.ReviewStatus != "pending" && *document.ReviewStatus != "confirmed") {
		return nil, apperr.Conflict("Este documento não pode ser confirmado.")
	}
	if *document.ReviewStatus == "confirmed" && document.LabReportID != nil {
		return s.report(ctx, *document.LabReportID)
	}
	result, err := s.drafts.Extraction(ctx, id)
	if err != nil {
		return nil, err
	}
	if !extraction.Usable(&result.Report) {
		return nil, apperr.DomainRuleViolation("A extração não contém resultados utilizáveis.")
	}
	report, err := mapExtractedToDomain(document.PatientID, document.UploadedByUserID, &result.Report)
	if err != nil {
		return nil, apperr.DomainRuleViolation("Os dados extraídos não permitem confirmar este exame.")
	}
	report.ExamDocumentID = &id
	reportID, err := s.repository.Confirm(ctx, id, userID, report, generateLabFingerprint(document.PatientID, report), result.Report.RawText)
	if errors.Is(err, domain.ErrReviewConflict) {
		return nil, apperr.Conflict("O rascunho foi descartado ou não pode ser confirmado.")
	}
	if errors.Is(err, labs.ErrLabReportAlreadyExists) {
		return nil, apperr.AlreadyExists("Este exame já está cadastrado em outro documento.")
	}
	if err != nil {
		return nil, apperr.Internal("Falha ao confirmar o exame.", err)
	}
	return s.report(ctx, reportID)
}

func (s *Service) report(ctx context.Context, id uuid.UUID) (*laboratory.LabReportOutput, error) {
	report, err := s.labs.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal("Falha ao consultar o exame confirmado.", err)
	}
	if report == nil {
		return nil, apperr.NotFound("Exame não encontrado.")
	}
	return toOutput(report), nil
}
