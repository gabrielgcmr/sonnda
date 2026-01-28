package labsuc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	labsvc "github.com/gabrielgcmr/sonnda/internal/app/services/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type CreateLabReportFromDocumentUseCase interface {
	Execute(ctx context.Context, input CreateLabReportFromDocumentInput) (*labsvc.LabReportOutput, error)
}

type createLabReportFromDocumentUseCase struct {
	patientRepo ports.PatientRepo
	labsRepo    ports.LabsRepo
	extractor   ports.DocumentExtractorService
}

var _ CreateLabReportFromDocumentUseCase = (*createLabReportFromDocumentUseCase)(nil)

func NewCreateLabReportFromDocument(
	patientRepo ports.PatientRepo,
	labsRepo ports.LabsRepo,
	extractor ports.DocumentExtractorService,
) CreateLabReportFromDocumentUseCase {
	return &createLabReportFromDocumentUseCase{
		patientRepo: patientRepo,
		labsRepo:    labsRepo,
		extractor:   extractor,
	}
}

func (u *createLabReportFromDocumentUseCase) Execute(ctx context.Context, input CreateLabReportFromDocumentInput) (*labsvc.LabReportOutput, error) {
	//Valida o input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	p, err := u.patientRepo.FindByID(ctx, input.PatientID)
	if err != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   err,
		}
	}
	if p == nil {
		return nil, &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "paciente não encontrado",
		}
	}

	extracted, err := u.extractor.ExtractLabReport(ctx, input.DocumentURI, input.MimeType)
	if err != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_EXTERNAL_SERVICE_ERROR,
			Message: "falha ao processar documento",
			Cause:   err,
		}
	}

	report, err := u.mapExtractedToDomain(input.PatientID, input.UploadedByUserID, extracted)
	if err != nil {
		return nil, u.mapDomainError(err)
	}

	fingerprint := generateLabFingerprint(input.PatientID, report)

	exists, err := u.labsRepo.ExistsBySignature(ctx, input.PatientID, fingerprint)
	if err != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   err,
		}
	}
	if exists {
		return nil, &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "laudo já existe",
		}
	}

	report.Fingerprint = &fingerprint
	if err := u.labsRepo.Create(ctx, report); err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr != nil {
			return nil, appErr
		}
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   err,
		}
	}

	return toOutput(report), nil
}

func (u *createLabReportFromDocumentUseCase) validateInput(input CreateLabReportFromDocumentInput) error {
	var violations []apperr.Violation

	if input.PatientID == uuid.Nil {
		violations = append(violations, apperr.Violation{Field: "patient_id", Reason: "required"})
	}
	if input.UploadedByUserID == uuid.Nil {
		violations = append(violations, apperr.Violation{Field: "uploaded_by_user_id", Reason: "required"})
	}
	if strings.TrimSpace(input.DocumentURI) == "" {
		violations = append(violations, apperr.Violation{Field: "document_uri", Reason: "required"})
	}

	switch strings.ToLower(strings.TrimSpace(input.MimeType)) {
	case "application/pdf", "image/jpeg", "image/png":
	default:
		violations = append(violations, apperr.Violation{Field: "mime_type", Reason: "unsupported"})
	}

	if len(violations) > 0 {
		return apperr.Validation("entrada inválida", violations...)
	}

	return nil
}

func (u *createLabReportFromDocumentUseCase) mapExtractedToDomain(
	patientID uuid.UUID,
	uploadedByUserID uuid.UUID,
	extracted *ports.ExtractedLabReport,
) (*labs.LabReport, error) {
	if extracted == nil {
		return nil, labs.ErrInvalidInput
	}

	report, err := labs.NewLabReport(patientID.String(), uploadedByUserID.String())
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
		if t, err := parseDate(*extracted.PatientDOB); err == nil {
			report.PatientDOB = &t
		}
	}
	if extracted.ReportDate != nil {
		if t, err := parseDate(*extracted.ReportDate); err == nil {
			report.ReportDate = &t
		}
	}

	for _, et := range extracted.Tests {
		testResult, err := labs.NewLabResult(report.ID.String(), et.TestName)
		if err != nil {
			return nil, err
		}

		testResult.Material = et.Material
		testResult.Method = et.Method

		if et.CollectedAt != nil {
			if t, err := parseDateTime(*et.CollectedAt); err == nil {
				testResult.CollectedAt = &t
			}
		}
		if et.ReleaseAt != nil {
			if t, err := parseDateTime(*et.ReleaseAt); err == nil {
				testResult.ReleaseAt = &t
			}
		}

		for _, ei := range et.Items {
			item, err := labs.NewLabResultItem(testResult.ID.String(), ei.ParameterName)
			if err != nil {
				return nil, err
			}
			item.ResultValue = ei.ResultValue
			item.ResultUnit = ei.ResultUnit
			item.ReferenceText = ei.ReferenceText
			item.Normalize()
			testResult.Items = append(testResult.Items, *item)
		}

		testResult.Normalize()
		report.TestResults = append(report.TestResults, *testResult)
	}

	report.Normalize()
	report.UpdatedAt = time.Now().UTC()

	return report, nil
}

func (u *createLabReportFromDocumentUseCase) mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return appErr
	}

	switch {
	case errors.Is(err, labs.ErrInvalidInput),
		errors.Is(err, labs.ErrMissingId),
		errors.Is(err, labs.ErrInvalidDateFormat),
		errors.Is(err, labs.ErrInvalidDocument),
		errors.Is(err, labs.ErrInvalidPatientID),
		errors.Is(err, labs.ErrInvalidUploadedByUser),
		errors.Is(err, labs.ErrInvalidTestName),
		errors.Is(err, labs.ErrInvalidParameterName):
		return &apperr.AppError{Code: apperr.VALIDATION_FAILED, Message: "entrada inválida", Cause: err}
	default:
		return &apperr.AppError{Code: apperr.INTERNAL_ERROR, Message: "erro inesperado", Cause: err}
	}
}

func parseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)

	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		"2006/01/02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}

	return time.Time{}, labs.ErrInvalidDateFormat
}

func parseDateTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\u00e0s", " ")
	raw = strings.ReplaceAll(raw, " as ", " ")
	raw = strings.ReplaceAll(raw, "h", ":")
	raw = strings.Join(strings.Fields(raw), " ")

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05-07:00",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02/01/2006 15:04:05 -0700",
		"02/01/2006 15:04 -0700",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}

	return parseDate(raw)
}

func normalize(s string) string {
	return strings.TrimSpace(strings.ToUpper(s))
}

func normalizeValue(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func generateLabFingerprint(patientID uuid.UUID, labReport *labs.LabReport) string {
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

func toOutput(report *labs.LabReport) *labsvc.LabReportOutput {
	output := &labsvc.LabReportOutput{
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
		testOutput := labsvc.TestResultOutput{
			ID:          tr.ID,
			TestName:    tr.TestName,
			Material:    tr.Material,
			Method:      tr.Method,
			CollectedAt: tr.CollectedAt,
			ReleaseAt:   tr.ReleaseAt,
		}

		for _, item := range tr.Items {
			testOutput.Items = append(testOutput.Items, labsvc.TestItemOutput{
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
