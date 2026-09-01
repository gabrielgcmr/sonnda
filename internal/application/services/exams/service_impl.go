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

func (s *service) CreateReportFromText(ctx context.Context, input CreateExamReportFromTextInput) (*ExamReportOutput, error) {
	if input.ExamDocumentID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_document_id", Reason: "required"})
	}
	if input.PatientID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}
	if input.UploadedByUserID == uuid.Nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "uploaded_by_user_id", Reason: "required"})
	}

	if existing, err := s.examsRepo.FindReportByDocumentID(ctx, input.ExamDocumentID); err != nil {
		return nil, mapRepoError("exams.find_report_by_document_id", err)
	} else if existing != nil {
		return mapDomainReportToOutput(existing), nil
	}

	report, err := exams.NewExamReport(&input.ExamDocumentID, input.PatientID, input.UploadedByUserID, input.Category, input.ReportText)
	if err != nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_report", Reason: err.Error()})
	}

	metadata := inferReportMetadata(input.ReportText)
	report.Title = metadata.Title
	report.Modality = metadata.Modality
	report.BodySite = metadata.BodySite
	report.Conclusion = metadata.Conclusion
	report.ExtractionMethod = stringToOptional(input.ExtractionMethod)
	report.Confidence = input.Confidence
	if err := report.NormalizeAndValidate(); err != nil {
		return nil, apperr.Validation("entrada invalida", apperr.Violation{Field: "exam_report", Reason: err.Error()})
	}

	if err := s.examsRepo.CreateReport(ctx, report); err != nil {
		return nil, mapRepoError("exams.create_report", err)
	}

	return mapDomainReportToOutput(report), nil
}

func (s *service) ListReportsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]ExamReportOutput, error) {
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

	reports, err := s.examsRepo.ListReportsByPatient(ctx, patientID, limit, offset)
	if err != nil {
		return nil, mapRepoError("exams.list_reports_by_patient", err)
	}

	out := make([]ExamReportOutput, 0, len(reports))
	for _, report := range reports {
		out = append(out, *mapDomainReportToOutput(&report))
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

func mapDomainReportToOutput(report *exams.ExamReport) *ExamReportOutput {
	return &ExamReportOutput{
		ID:                 report.ID,
		ExamDocumentID:     report.ExamDocumentID,
		PatientID:          report.PatientID,
		UploadedByUserID:   report.UploadedByUserID,
		Category:           report.Category,
		Title:              report.Title,
		Modality:           report.Modality,
		BodySite:           report.BodySite,
		PerformedAt:        report.PerformedAt,
		FacilityName:       report.FacilityName,
		InterpretingDoctor: report.InterpretingDoctor,
		ReportText:         report.ReportText,
		Conclusion:         report.Conclusion,
		ExtractionMethod:   report.ExtractionMethod,
		Confidence:         report.Confidence,
		CreatedAt:          report.CreatedAt,
		UpdatedAt:          report.UpdatedAt,
	}
}

func stringToOptional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
