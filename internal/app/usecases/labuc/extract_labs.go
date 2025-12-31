package labuc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"sonnda-api/internal/domain/entities/medicalrecord/lab"
	"sonnda-api/internal/domain/ports/integrations"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

type ExtractLabsUseCase struct {
	repo      repositories.LabRepository
	extractor integrations.DocumentExtractor
}

func NewExtractLabs(
	repo repositories.LabRepository,
	extractor integrations.DocumentExtractor,
) *ExtractLabsUseCase {
	return &ExtractLabsUseCase{
		repo:      repo,
		extractor: extractor,
	}
}

// CreateFromDocument:
// 1) chama o extractor (Document AI via adapter de labs)
// 2) converte DTO -> dominio
// 3) salva no banco via Repository
func (uc *ExtractLabsUseCase) Execute(
	ctx context.Context,
	input CreateFromDocumentInput,
) (*LabReportOutput, error) {
	if err := uc.validate(input); err != nil {
		return nil, err
	}

	extracted, err := uc.extractor.ExtractLabReport(ctx, input.DocumentURI, input.MimeType)
	if err != nil {
		return nil, lab.ErrDocumentProcessing
	}

	report, err := uc.mapExtractedToDomain(input.PatientID, input.UploadedByUserID, extracted)
	if err != nil {
		return nil, err
	}

	fingerprint := generateLabFingerprint(input.PatientID, report)

	exists, err := uc.repo.ExistsBySignature(ctx, input.PatientID, fingerprint)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, lab.ErrLabReportAlreadyExists
	}

	report.Fingerprint = &fingerprint
	if err := uc.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	return uc.toOutput(report), nil
}

func (uc *ExtractLabsUseCase) validate(input CreateFromDocumentInput) error {
	if input.PatientID == uuid.Nil {
		return lab.ErrInvalidInput
	}
	if input.UploadedByUserID == uuid.Nil {
		return lab.ErrInvalidUploadedByUser
	}
	if strings.TrimSpace(input.DocumentURI) == "" {
		return lab.ErrInvalidDocument
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
		return lab.ErrInvalidDocument
	}

	return nil
}

func (uc *ExtractLabsUseCase) mapExtractedToDomain(
	patientID uuid.UUID,
	uploadedByUserID uuid.UUID,
	extracted *integrations.ExtractedLabReport,
) (*lab.LabReport, error) {
	if extracted == nil {
		return nil, lab.ErrInvalidInput
	}

	report, err := lab.NewLabReport(patientID.String(), uploadedByUserID.String())
	if err != nil {
		return nil, err
	}

	report.PatientName = extracted.PatientName
	report.LabName = extracted.LabName
	report.LabPhone = extracted.LabPhone
	report.InsuranceProvider = extracted.InsuranceProvider
	report.RequestingDoctor = extracted.RequestingDoctor
	report.TechnicalManager = extracted.TechnicalManager
	report.RawText = extracted.RawText

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

	for _, et := range extracted.Tests {
		result, err := lab.NewLabResult(report.ID.String(), et.TestName)
		if err != nil {
			return nil, err
		}

		result.Material = et.Material
		result.Method = et.Method

		if et.CollectedAt != nil {
			if t, err := uc.parseDateTime(*et.CollectedAt); err == nil {
				result.CollectedAt = &t
			}
		}
		if et.ReleaseAt != nil {
			if t, err := uc.parseDateTime(*et.ReleaseAt); err == nil {
				result.ReleaseAt = &t
			}
		}

		for _, ei := range et.Items {
			item, err := lab.NewLabResultItem(result.ID.String(), ei.ParameterName)
			if err != nil {
				return nil, err
			}
			item.ResultValue = ei.ResultValue
			item.ResultUnit = ei.ResultUnit
			item.ReferenceText = ei.ReferenceText
			item.Normalize()
			result.Items = append(result.Items, *item)
		}

		result.Normalize()
		report.TestResults = append(report.TestResults, *result)
	}

	report.Normalize()
	report.UpdatedAt = time.Now().UTC()

	return report, nil
}

// Gera a assinatura do laudo com base nos dados chave
func generateLabFingerprint(patientID uuid.UUID, labReport *lab.LabReport) string {
	var parts []string
	patientKey := patientID.String()
	if patientKey == "" && labReport != nil {
		patientKey = labReport.PatientID.String()
	}

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
				fmt.Sprintf("%s|%s|%s|%s|%s", patientKey, dateStr, testName, param, value),
			)
		}
	}

	sort.Strings(parts)

	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func (uc *ExtractLabsUseCase) toOutput(report *lab.LabReport) *LabReportOutput {
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
		UploadedByUserID:  report.UploadedBy,
		Fingerprint:       report.Fingerprint,
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
