package labs

import (
	"context"
	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
)

type ListFullLabsUseCase struct {
	patientRepo repositories.PatientRepository
	labsRepo    repositories.LabsRepository
}

func NewListFullLabs(
	patientRepo repositories.PatientRepository,
	labsRepo repositories.LabsRepository,
) *ListFullLabsUseCase {
	return &ListFullLabsUseCase{
		patientRepo: patientRepo,
		labsRepo:    labsRepo,
	}
}

func (uc *ListFullLabsUseCase) Execute(
	ctx context.Context,
	patientID string,
	limit, offset int,
) ([]*LabReportOutput, error) {
	if patientID == "" {
		return nil, domain.ErrInvalidInput
	}

	patient, err := uc.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	headers, err := uc.labsRepo.FindByPatientID(ctx, patient.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]*LabReportOutput, 0, len(headers))

	for _, header := range headers {
		fullReport, err := uc.labsRepo.FindByID(ctx, header.ID)
		if err != nil {
			return nil, err
		}
		if fullReport == nil {
			continue
		}

		// aqui você pode reutilizar a lógica de toOutput do ExtractLabsUseCase
		// idealmente extraindo isso pra uma função pura do pacote labs
		dto := mapDomainReportToOutput(fullReport)
		out = append(out, dto)
	}

	return out, nil
}

func mapDomainReportToOutput(report *domain.LabReport) *LabReportOutput {
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
