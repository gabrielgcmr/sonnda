// internal/features/documentprocessing/service_impl.go
package documentprocessing

import (
	"context"

	exams "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type service struct {
	patientRepo patientprofile.Repository
	examsRepo   DocumentRepository
}

var _ Service = (*service)(nil)

func New(patientRepo patientprofile.Repository, examsRepo DocumentRepository) Service {
	return &service{
		patientRepo: patientRepo,
		examsRepo:   examsRepo,
	}
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

func (s *service) ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamDocumentTextOutput, error) {
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

	documentTexts, err := s.examsRepo.ListDocumentTextsByPatient(ctx, patientID, limit, offset)
	if err != nil {
		return nil, mapRepoError("exams.list_document_texts_by_patient", err)
	}

	out := make([]ExamDocumentTextOutput, 0, len(documentTexts))
	for _, documentText := range documentTexts {
		out = append(out, *mapDomainDocumentTextToOutput(&documentText))
	}

	return out, nil
}

func mapDomainDocumentToOutput(document *exams.ExamDocument) *ExamDocumentOutput {
	return &ExamDocumentOutput{
		ReviewStatus: document.ReviewStatus, LabReportID: document.LabReportID, ConfirmedByUserID: document.ConfirmedByUserID, ConfirmedAt: document.ConfirmedAt,
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

func mapDomainDocumentTextToOutput(documentText *exams.ExamDocumentText) *ExamDocumentTextOutput {
	return &ExamDocumentTextOutput{
		ID:                 documentText.ID,
		ExamDocumentID:     documentText.ExamDocumentID,
		PatientID:          documentText.PatientID,
		UploadedByUserID:   documentText.UploadedByUserID,
		Category:           documentText.Category,
		Title:              documentText.Title,
		Modality:           documentText.Modality,
		BodySite:           documentText.BodySite,
		PerformedAt:        documentText.PerformedAt,
		FacilityName:       documentText.FacilityName,
		InterpretingDoctor: documentText.InterpretingDoctor,
		Text:               documentText.Text,
		Conclusion:         documentText.Conclusion,
		ExtractionMethod:   documentText.ExtractionMethod,
		Confidence:         documentText.Confidence,
		CreatedAt:          documentText.CreatedAt,
		UpdatedAt:          documentText.UpdatedAt,
	}
}
