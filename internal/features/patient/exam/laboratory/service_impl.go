// internal/features/patient/exam/laboratory/service_impl.go
package laboratory

import (
	"context"

	labdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type service struct {
	patientRepo patientprofile.Repository
	labsRepo    Repository
}

var _ Service = (*service)(nil)

func New(
	patientRepo patientprofile.Repository,
	labsRepo Repository,
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
			ID:             fullReport.ID,
			PatientID:      fullReport.PatientID,
			ExamDocumentID: fullReport.ExamDocumentID,
			ReportDate:     fullReport.ReportDate,
		}

		for _, tr := range fullReport.TestResults {
			panelSummary := LabPanelSummaryOutput{
				TestName:    tr.TestName,
				CollectedAt: tr.CollectedAt,
			}

			for _, item := range tr.Items {
				panelSummary.Observations = append(panelSummary.Observations, ObservationSummaryOutput{
					ParameterName: item.ParameterName,
					ResultValue:   item.ResultValue,
					ResultUnit:    item.ResultUnit,
				})
			}

			summary.SummaryPanels = append(summary.SummaryPanels, panelSummary)
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

func (s *service) FindByID(ctx context.Context, reportID uuid.UUID) (*LabReportOutput, error) {
	if reportID == uuid.Nil {
		return nil, apperr.Validation("entrada inválida", apperr.Violation{Field: "id", Reason: "required"})
	}

	report, err := s.labsRepo.FindByID(ctx, reportID)
	if err != nil {
		return nil, mapRepoError("labs.find_by_id", err)
	}
	if report == nil {
		return nil, nil
	}

	return mapDomainReportToOutput(report), nil
}

func mapDomainReportToOutput(report *labdomain.LabReport) *LabReportOutput {
	output := &LabReportOutput{
		ID:                report.ID,
		PatientID:         report.PatientID,
		ExamDocumentID:    report.ExamDocumentID,
		PatientName:       report.PatientName,
		PatientDOB:        report.PatientDOB,
		LabName:           report.LabName,
		LabPhone:          report.LabPhone,
		InsuranceProvider: report.InsuranceProvider,
		RequestingDoctor:  report.RequestingDoctor,
		TechnicalManager:  report.TechnicalManager,
		ReportDate:        report.ReportDate,
		UploadedByUserID:  report.UploadedBy,
		CreatedAt:         report.CreatedAt,
		UpdatedAt:         report.UpdatedAt,
	}

	for _, tr := range report.TestResults {
		panelOutput := LabPanelOutput{
			ID:          tr.ID,
			TestName:    tr.TestName,
			Material:    tr.Material,
			Method:      tr.Method,
			CollectedAt: tr.CollectedAt,
			ReleaseAt:   tr.ReleaseAt,
		}

		for _, item := range tr.Items {
			panelOutput.Observations = append(panelOutput.Observations, ObservationOutput{
				ID:            item.ID,
				ParameterName: item.ParameterName,
				ResultValue:   item.ResultValue,
				ResultUnit:    item.ResultUnit,
				ReferenceText: item.ReferenceText,
			})
		}

		output.Panels = append(output.Panels, panelOutput)
	}

	return output
}
