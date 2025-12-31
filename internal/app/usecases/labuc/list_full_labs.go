package labuc

import (
	"context"

	"sonnda-api/internal/domain/entities/medicalrecord/lab"
	"sonnda-api/internal/domain/entities/patient"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

type ListFullLabsUseCase struct {
	patientRepo repositories.PatientRepository
	labsRepo    repositories.LabRepository
}

func NewListFullLabs(
	patientRepo repositories.PatientRepository,
	labsRepo repositories.LabRepository,
) *ListFullLabsUseCase {
	return &ListFullLabsUseCase{
		patientRepo: patientRepo,
		labsRepo:    labsRepo,
	}
}

func (uc *ListFullLabsUseCase) Execute(
	ctx context.Context,
	patientID uuid.UUID,
	limit, offset int,
) ([]*LabReportOutput, error) {
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

	headers, err := uc.labsRepo.ListLabs(ctx, p.ID, limit, offset)
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

		dto := mapDomainReportToOutput(fullReport)
		out = append(out, dto)
	}

	return out, nil
}

func mapDomainReportToOutput(report *lab.LabReport) *LabReportOutput {
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
