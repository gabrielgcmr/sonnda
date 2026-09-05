// internal/application/services/exams/service_impl.go
package examsvc

import (
	"context"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type service struct {
	patientRepo repository.Patient
	examsRepo   repository.Exams
	router      ExamRouter
}

var _ Service = (*service)(nil)

func New(patientRepo repository.Patient, examsRepo repository.Exams) Service {
	return &service{
		patientRepo: patientRepo,
		examsRepo:   examsRepo,
		router:      NewHeuristicExamRouter(),
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

func (s *service) RouteDocument(ctx context.Context, input RouteExamDocumentInput) (*ExamDocumentOutput, error) {
	if input.ID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "id", Reason: "required"})
	}

	document, err := s.examsRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, mapRepoError("exams.find_by_id", err)
	}
	if document == nil {
		return nil, nil
	}

	extractionMethod := strings.TrimSpace(input.ExtractionMethod)
	if extractionMethod == "" {
		extractionMethod = "unknown"
	}

	route := s.router.Route(ExamRouteInput{
		ExtractedText:    input.ExtractedText,
		MimeType:         document.MimeType,
		OriginalFilename: document.OriginalFilename,
	})

	updated, err := s.examsRepo.MarkClassified(
		ctx,
		input.ID,
		route.Status,
		route.ExamType,
		&extractionMethod,
		&route.Confidence,
		&input.ExtractedText,
	)
	if err != nil {
		return nil, mapRepoError("exams.mark_classified", err)
	}
	if updated == nil {
		return nil, nil
	}

	return mapDomainDocumentToOutput(updated), nil
}

func (s *service) MarkFailed(ctx context.Context, input MarkExamDocumentFailedInput) (*ExamDocumentOutput, error) {
	if input.ID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "id", Reason: "required"})
	}

	errorMessage := strings.TrimSpace(input.ErrorMessage)
	if errorMessage == "" {
		errorMessage = "falha ao processar exame"
	}

	document, err := s.examsRepo.MarkFailed(ctx, input.ID, errorMessage)
	if err != nil {
		return nil, mapRepoError("exams.mark_failed", err)
	}
	if document == nil {
		return nil, nil
	}

	return mapDomainDocumentToOutput(document), nil
}

func (s *service) CreateDocumentTextFromText(ctx context.Context, input CreateExamDocumentTextFromTextInput) (*ExamDocumentTextOutput, error) {
	if input.ExamDocumentID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_document_id", Reason: "required"})
	}
	if input.PatientID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}
	if input.UploadedByUserID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "uploaded_by_user_id", Reason: "required"})
	}

	if existing, err := s.examsRepo.FindDocumentTextByDocumentID(ctx, input.ExamDocumentID); err != nil {
		return nil, mapRepoError("exams.find_document_text_by_document_id", err)
	} else if existing != nil {
		return mapDomainDocumentTextToOutput(existing), nil
	}

	documentText, err := exams.NewExamDocumentText(&input.ExamDocumentID, input.PatientID, input.UploadedByUserID, input.Category, input.Text)
	if err != nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_document_text", Reason: err.Error()})
	}

	metadata := inferDocumentTextMetadata(input.Text)
	documentText.Title = metadata.Title
	documentText.Modality = metadata.Modality
	documentText.BodySite = metadata.BodySite
	documentText.Conclusion = metadata.Conclusion
	documentText.ExtractionMethod = stringToOptional(input.ExtractionMethod)
	documentText.Confidence = input.Confidence
	if err := documentText.NormalizeAndValidate(); err != nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_document_text", Reason: err.Error()})
	}

	if err := s.examsRepo.CreateDocumentText(ctx, documentText); err != nil {
		return nil, mapRepoError("exams.create_document_text", err)
	}

	return mapDomainDocumentTextToOutput(documentText), nil
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

func stringToOptional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
