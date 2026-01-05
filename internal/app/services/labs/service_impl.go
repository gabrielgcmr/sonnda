package labsvc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"sonnda-api/internal/app/interfaces/external"
	"sonnda-api/internal/app/interfaces/repositories"
	applog "sonnda-api/internal/app/observability"
	"sonnda-api/internal/domain/model/medicalrecord/labs"
	"sonnda-api/internal/domain/model/patient"

	"github.com/google/uuid"
)

type service struct {
	patientRepo repositories.PatientRepository
	labsRepo    repositories.LabRepository
	extractor   external.DocumentExtractor
}

var _ Service = (*service)(nil)

func New(
	patientRepo repositories.PatientRepository,
	labsRepo repositories.LabRepository,
	extractor external.DocumentExtractor,
) Service {
	return &service{
		patientRepo: patientRepo,
		labsRepo:    labsRepo,
		extractor:   extractor,
	}
}

func (s *service) CreateFromDocument(ctx context.Context, input CreateFromDocumentInput) (*LabReportOutput, error) {
	if err := s.validateCreate(input); err != nil {
		return nil, err
	}

	extracted, err := s.extractor.ExtractLabReport(ctx, input.DocumentURI, input.MimeType)
	if err != nil {
		applog.FromContext(ctx).Warn("extract_lab_report_failed", "err", err)
		return nil, labs.ErrDocumentProcessing
	}

	report, err := s.mapExtractedToDomain(input.PatientID, input.UploadedByUserID, extracted)
	if err != nil {
		return nil, err
	}

	fingerprint := generateLabFingerprint(input.PatientID, report)

	exists, err := s.labsRepo.ExistsBySignature(ctx, input.PatientID, fingerprint)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, labs.ErrLabReportAlreadyExists
	}

	report.Fingerprint = &fingerprint
	if err := s.labsRepo.Create(ctx, report); err != nil {
		return nil, err
	}

	return s.toOutput(report), nil
}

func (s *service) List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]LabReportSummaryOutput, error) {
	if patientID == uuid.Nil {
		return nil, labs.ErrInvalidInput
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	reports, err := s.labsRepo.ListLabs(ctx, p.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]LabReportSummaryOutput, 0, len(reports))

	for _, header := range reports {
		fullReport, err := s.labsRepo.FindByID(ctx, header.ID)
		if err != nil {
			return nil, err
		}
		if fullReport == nil {
			continue
		}

		summary := LabReportSummaryOutput{
			ID:         fullReport.ID,
			PatientID:  fullReport.PatientID,
			ReportDate: fullReport.ReportDate,
		}

		for _, tr := range fullReport.TestResults {
			testSummary := LabResultSummaryOutput{
				TestName:    tr.TestName,
				CollectedAt: tr.CollectedAt,
			}

			for _, item := range tr.Items {
				testSummary.Items = append(testSummary.Items, ResultItemSummaryOutput{
					ParameterName: item.ParameterName,
					ResultValue:   item.ResultValue,
					ResultUnit:    item.ResultUnit,
				})
			}

			summary.SummaryTests = append(summary.SummaryTests, testSummary)
		}

		out = append(out, summary)
	}

	return out, nil
}

func (s *service) ListFull(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LabReportOutput, error) {
	if patientID == uuid.Nil {
		return nil, labs.ErrInvalidInput
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	headers, err := s.labsRepo.ListLabs(ctx, p.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]*LabReportOutput, 0, len(headers))

	for _, header := range headers {
		fullReport, err := s.labsRepo.FindByID(ctx, header.ID)
		if err != nil {
			return nil, err
		}
		if fullReport == nil {
			continue
		}

		dto := mapDomainReportToOutput(fullReport)
		out = append(out, dto)
	}

	return out, nil
}

func (s *service) validateCreate(input CreateFromDocumentInput) error {
	if input.PatientID == uuid.Nil {
		return labs.ErrInvalidInput
	}
	if input.UploadedByUserID == uuid.Nil {
		return labs.ErrInvalidUploadedByUser
	}
	if strings.TrimSpace(input.DocumentURI) == "" {
		return labs.ErrInvalidDocument
	}

	validMimeTypes := []string{"application/pdf", "image/jpeg", "image/png"}
	for _, mt := range validMimeTypes {
		if input.MimeType == mt {
			return nil
		}
	}

	return labs.ErrInvalidDocument
}

func (s *service) mapExtractedToDomain(
	patientID uuid.UUID,
	uploadedByUserID uuid.UUID,
	extracted *external.ExtractedLabReport,
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
		if t, err := s.parseDate(*extracted.PatientDOB); err == nil {
			report.PatientDOB = &t
		}
	}
	if extracted.ReportDate != nil {
		if t, err := s.parseDate(*extracted.ReportDate); err == nil {
			report.ReportDate = &t
		}
	}

	for _, et := range extracted.Tests {
		result, err := labs.NewLabResult(report.ID.String(), et.TestName)
		if err != nil {
			return nil, err
		}

		result.Material = et.Material
		result.Method = et.Method

		if et.CollectedAt != nil {
			if t, err := s.parseDateTime(*et.CollectedAt); err == nil {
				result.CollectedAt = &t
			}
		}
		if et.ReleaseAt != nil {
			if t, err := s.parseDateTime(*et.ReleaseAt); err == nil {
				result.ReleaseAt = &t
			}
		}

		for _, ei := range et.Items {
			item, err := labs.NewLabResultItem(result.ID.String(), ei.ParameterName)
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

func (s *service) parseDate(raw string) (time.Time, error) {
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

func (s *service) parseDateTime(raw string) (time.Time, error) {
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

	return s.parseDate(raw)
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

func (s *service) toOutput(report *labs.LabReport) *LabReportOutput {
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

func mapDomainReportToOutput(report *labs.LabReport) *LabReportOutput {
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
