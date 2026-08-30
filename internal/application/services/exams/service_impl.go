package examsvc

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type service struct {
	patientRepo repository.Patient
	examsRepo   repository.Exams
}

var _ Service = (*service)(nil)

func New(patientRepo repository.Patient, examsRepo repository.Exams) Service {
	return &service{
		patientRepo: patientRepo,
		examsRepo:   examsRepo,
	}
}

func (s *service) Create(ctx context.Context, input CreateExamDocumentInput) (*ExamDocumentOutput, error) {
	if input.PatientID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}
	if input.UploadedByUserID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "uploaded_by_user_id", Reason: "required"})
	}

	p, err := s.patientRepo.FindByID(ctx, input.PatientID)
	if err != nil {
		return nil, mapRepoError("patient.find_by_id", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}

	document, err := exams.NewExamDocument(
		input.PatientID,
		input.UploadedByUserID,
		input.StorageURI,
		input.OriginalFilename,
		input.MimeType,
	)
	if err != nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_document", Reason: err.Error()})
	}

	if err := s.examsRepo.Create(ctx, document); err != nil {
		return nil, mapRepoError("exams.create", err)
	}

	return mapDomainDocumentToOutput(document), nil
}

func (s *service) FindByID(ctx context.Context, id uuid.UUID) (*ExamDocumentOutput, error) {
	if id == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "id", Reason: "required"})
	}

	document, err := s.examsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError("exams.find_by_id", err)
	}
	if document == nil {
		return nil, nil
	}

	return mapDomainDocumentToOutput(document), nil
}

func (s *service) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentOutput, error) {
	if patientID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, mapRepoError("patient.find_by_id", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}

	documents, err := s.examsRepo.ListByPatient(ctx, patientID, limit, offset)
	if err != nil {
		return nil, mapRepoError("exams.list_by_patient", err)
	}

	out := make([]ExamDocumentOutput, 0, len(documents))
	for _, document := range documents {
		out = append(out, *mapDomainDocumentToOutput(&document))
	}

	return out, nil
}

func mapDomainDocumentToOutput(document *exams.ExamDocument) *ExamDocumentOutput {
	return &ExamDocumentOutput{
		ID:               document.ID,
		PatientID:        document.PatientID,
		UploadedByUserID: document.UploadedByUserID,
		StorageURI:       document.StorageURI,
		OriginalFilename: document.OriginalFilename,
		MimeType:         document.MimeType,
		Status:           document.Status,
		ExamType:         document.ExamType,
		ExtractionMethod: document.ExtractionMethod,
		Confidence:       document.Confidence,
		ErrorMessage:     document.ErrorMessage,
		CreatedAt:        document.CreatedAt,
		UpdatedAt:        document.UpdatedAt,
	}
}
