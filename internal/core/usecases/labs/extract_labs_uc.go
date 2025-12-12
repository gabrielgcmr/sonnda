package labs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
	"sonnda-api/internal/core/ports/services"
)

type ExtractLabsUseCase struct {
	repo      repositories.LabsRepository
	extractor services.DocumentExtractor
}

func NewExtractLabs(
	repo repositories.LabsRepository,
	extractor services.DocumentExtractor,
) *ExtractLabsUseCase {
	return &ExtractLabsUseCase{
		repo:      repo,
		extractor: extractor,
	}
}

// CreateFromDocument:
// 1) chama o extractor (Document AI via adapter de labtest)
// 2) converte DTO -> dominio
// 3) salva no banco via Repository
func (uc *ExtractLabsUseCase) Execute(
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
	//4. Gera o identificador com o hash para ver duplicidade
	fingerprint := generateLabFingerprint(input.PatientID, report)

	// 5. checa existencia de duplicidade
	exists, err := uc.repo.ExistsBySignature(ctx, input.PatientID, fingerprint)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrLabReportAlreadyExists
	}

	// 5. Persiste no banco
	report.Fingerprint = &fingerprint
	if err := uc.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	// 6. Retorna output
	return uc.toOutput(report), nil
}

func (uc *ExtractLabsUseCase) validate(input CreateFromDocumentInput) error {
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

func (uc *ExtractLabsUseCase) mapExtractedToDomain(
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

// Gera a assinatura do laudo com base nos dados chave
func generateLabFingerprint(patientID string, labReport *domain.LabReport) string {
	var parts []string

	for _, tr := range labReport.TestResults {

		var dateStr string
		if tr.CollectedAt != nil {
			dateStr = tr.CollectedAt.Format("2006-01-02")
		} else if labReport.ReportDate != nil {
			dateStr = labReport.ReportDate.Format("2006-01-02")
		} else {
			dateStr = "000-00-00"
		}

		testName := normalize(tr.TestName)
		for _, item := range tr.Items {
			param := normalize(item.ParameterName)
			value := normalizeValue(item.ResultValue)

			parts = append(parts,
				fmt.Sprintf("%s|%s|%s|%s|%s", patientID, dateStr, testName, param, value),
			)
		}
	}

	sort.Strings(parts)

	// Cria o hash SHA-256
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)

}

func (uc *ExtractLabsUseCase) toOutput(report *domain.LabReport) *LabReportOutput {
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
