package labs

import (
	"context"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/google/uuid"
)

type ListLabsUseCase struct {
	patientRepo repositories.PatientRepository
	labsRepo    repositories.LabsRepository
}

func NewListLabs(
	patientRepo repositories.PatientRepository,
	labsRepo repositories.LabsRepository,
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
		return nil, domain.ErrInvalidInput
	}

	patient, err := uc.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	reports, err := uc.labsRepo.FindByPatientID(ctx, patient.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]LabReportSummaryOutput, 0, len(reports))

	for _, header := range reports {
		// 🔎 carrega o laudo completo pra poder montar SummaryTests
		fullReport, err := uc.labsRepo.FindByID(ctx, header.ID)
		if err != nil {
			return nil, err
		}
		if fullReport == nil {
			continue
		}

		summary := LabReportSummaryOutput{
			ID:         fullReport.ID.String(),
			PatientID:  fullReport.PatientID.String(),
			ReportDate: fullReport.ReportDate,
			// vamos preencher SummaryTests abaixo
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
