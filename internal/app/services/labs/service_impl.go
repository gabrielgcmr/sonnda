package labsvc

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/app/apperr"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"

	"github.com/google/uuid"
)

type service struct {
	patientRepo ports.PatientRepo
	labsRepo    ports.LabsRepo
}

var _ Service = (*service)(nil)

func New(
	patientRepo ports.PatientRepo,
	labsRepo ports.LabsRepo,
) Service {
	return &service{
		patientRepo: patientRepo,
		labsRepo:    labsRepo,
	}
}

func (s *service) List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]LabReportSummaryOutput, error) {
	if patientID == uuid.Nil {
		return nil, apperr.Validation("entrada inválida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, mapRepoError("patient.find_by_id", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}

	reports, err := s.labsRepo.ListLabs(ctx, p.ID, limit, offset)
	if err != nil {
		return nil, mapRepoError("labs.list", err)
	}

	out := make([]LabReportSummaryOutput, 0, len(reports))

	for _, header := range reports {
		fullReport, err := s.labsRepo.FindByID(ctx, header.ID)
		if err != nil {
			return nil, mapRepoError("labs.find_by_id", err)
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
		return nil, apperr.Validation("entrada inválida", apperr.Violation{Field: "patient_id", Reason: "required"})
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, mapRepoError("patient.find_by_id", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}

	headers, err := s.labsRepo.ListLabs(ctx, p.ID, limit, offset)
	if err != nil {
		return nil, mapRepoError("labs.list", err)
	}

	out := make([]*LabReportOutput, 0, len(headers))

	for _, header := range headers {
		fullReport, err := s.labsRepo.FindByID(ctx, header.ID)
		if err != nil {
			return nil, mapRepoError("labs.find_by_id", err)
		}
		if fullReport == nil {
			continue
		}

		dto := mapDomainReportToOutput(fullReport)
		out = append(out, dto)
	}

	return out, nil
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
