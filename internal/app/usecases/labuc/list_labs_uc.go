package labuc

import (
	"context"

	"sonnda-api/internal/domain/entities/medicalrecord/lab"
	"sonnda-api/internal/domain/entities/patient"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

type ListLabsUseCase struct {
	patientRepo repositories.PatientRepository
	labsRepo    repositories.LabRepository
}

func NewListLabs(
	patientRepo repositories.PatientRepository,
	labsRepo repositories.LabRepository,
) *ListLabsUseCase {
	return &ListLabsUseCase{
		patientRepo: patientRepo,
		labsRepo:    labsRepo,
	}
}

func (uc *ListLabsUseCase) Execute(
	ctx context.Context,
	patientID uuid.UUID,
	limit, offset int,
) ([]LabReportSummaryOutput, error) {
	if patientID == uuid.Nil {
		return nil, lab.ErrInvalidInput
	}

	p, err := uc.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	reports, err := uc.labsRepo.ListLabs(ctx, p.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]LabReportSummaryOutput, 0, len(reports))

	for _, header := range reports {
		fullReport, err := uc.labsRepo.FindByID(ctx, header.ID)
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
