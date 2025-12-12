package labs

import (
	"context"
	"time"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
	"sonnda-api/internal/core/ports/services"
)

type CreateFromDocumentUseCase struct {
	repo      repositories.LabsRepository
	extractor services.DocumentExtractor
}

func NewCreateFromDocument(
	repo repositories.LabsRepository,
	extractor services.DocumentExtractor,
) *CreateFromDocumentUseCase {
	return &CreateFromDocumentUseCase{
		repo:      repo,
		extractor: extractor,
	}
}

// CreateFromDocument:
// 1) chama o extractor (Document AI via adapter de labtest)
// 2) converte DTO -> dominio
// 3) salva no banco via Repository
func (uc *CreateFromDocumentUseCase) Execute(
	ctx context.Context,
	input CreateFromDocumentInput,
) (*LabReportOutput, error) {
	// 1. Validacoes basicas
	if err := uc.validate(input); err != nil {
		return nil, err
	}

	// 2. Extrai dados do documento via Document AI
	extracted, err := uc.extractor.ExtractLabReport(ctx, input.DocumentURI, input.MimeType)
	if err != nil {
		return nil, domain.ErrDocumentProcessing
	}

	// 3. Converte ExtractedLabReport -> Domain Entity
	var uploader *string
	if input.UploadedByUserID != "" {
		uploader = &input.UploadedByUserID
	}
	report, err := uc.mapExtractedToDomain(input.PatientID, uploader, extracted)
	if err != nil {
		return nil, err
	}

	// 4. Persiste no banco
	if err := uc.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	// 5. Retorna output
	return uc.toOutput(report), nil
}

func (uc *CreateFromDocumentUseCase) validate(input CreateFromDocumentInput) error {
	if input.PatientID == "" {
		return domain.ErrInvalidInput
	}
	if input.DocumentURI == "" {
		return domain.ErrInvalidDocument
	}

	validMimeTypes := []string{"application/pdf", "image/jpeg", "image/png"}
	isValid := false
	for _, mt := range validMimeTypes {
		if input.MimeType == mt {
			isValid = true
			break
		}
	}
	if !isValid {
		return domain.ErrInvalidDocument
	}

	return nil
}

func (uc *CreateFromDocumentUseCase) mapExtractedToDomain(
	patientID string,
	uploadedByUserID *string,
	extracted *services.ExtractedLabReport,
) (*domain.LabReport, error) {
	now := time.Now()

	report := &domain.LabReport{
		PatientID:         patientID,
		PatientName:       extracted.PatientName,
		LabName:           extracted.LabName,
		LabPhone:          extracted.LabPhone,
		InsuranceProvider: extracted.InsuranceProvider,
		RequestingDoctor:  extracted.RequestingDoctor,
		TechnicalManager:  extracted.TechnicalManager,
		RawText:           extracted.RawText,
		CreatedAt:         now,
		UpdatedAt:         now,
		UploadedByUserID:  uploadedByUserID,
	}

	// Parse datas
	if extracted.PatientDOB != nil {
		if t, err := uc.parseDate(*extracted.PatientDOB); err == nil {
			report.PatientDOB = &t
		}
	}
	if extracted.ReportDate != nil {
		if t, err := uc.parseDate(*extracted.ReportDate); err == nil {
			report.ReportDate = &t
		}
	}

	// Mapeia testes
	for _, et := range extracted.Tests {
		testResult := domain.LabTestResult{
			TestName: et.TestName,
			Material: et.Material,
			Method:   et.Method,
		}

		// Parse datas do teste
		if et.CollectedAt != nil {
			if t, err := uc.parseDateTime(*et.CollectedAt); err == nil {
				testResult.CollectedAt = &t
			}
		}
		if et.ReleaseAt != nil {
			if t, err := uc.parseDateTime(*et.ReleaseAt); err == nil {
				testResult.ReleaseAt = &t
			}
		}

		// Mapeia itens
		for _, ei := range et.Items {
			item := domain.LabTestItem{
				ParameterName: ei.ParameterName,
				ResultValue:   ei.ResultValue,
				ResultUnit:    ei.ResultUnit,
				ReferenceText: ei.ReferenceText,
			}
			testResult.Items = append(testResult.Items, item)
		}

		report.TestResults = append(report.TestResults, testResult)
	}

	return report, nil
}

func (uc *CreateFromDocumentUseCase) toOutput(report *domain.LabReport) *LabReportOutput {
	output := &LabReportOutput{
		ID:                report.ID,
		PatientID:         report.PatientID,
		PatientName:       report.PatientName,
		PatientDOB:        report.PatientDOB,
		LabName:           report.LabName,
		LabPhone:          report.LabPhone,
		InsuranceProvider: report.InsuranceProvider,
		RequestingDoctor:  report.RequestingDoctor,
		TechnicalManager:  report.TechnicalManager,
		ReportDate:        report.ReportDate,
		UploadedByUserID:  report.UploadedByUserID,
		CreatedAt:         report.CreatedAt,
		UpdatedAt:         report.UpdatedAt,
	}

	for _, tr := range report.TestResults {
		testOutput := TestResultOutput{
			ID:          tr.ID,
			TestName:    tr.TestName,
			Material:    tr.Material,
			Method:      tr.Method,
			CollectedAt: tr.CollectedAt,
			ReleaseAt:   tr.ReleaseAt,
		}

		for _, item := range tr.Items {
			testOutput.Items = append(testOutput.Items, TestItemOutput{
				ID:            item.ID,
				ParameterName: item.ParameterName,
				ResultValue:   item.ResultValue,
				ResultUnit:    item.ResultUnit,
				ReferenceText: item.ReferenceText,
			})
		}

		output.TestResults = append(output.TestResults, testOutput)
	}

	return output
}
