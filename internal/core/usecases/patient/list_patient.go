package patient

import (
	"context"
	"sonnda-api/internal/core/domain/patient"
	"sonnda-api/internal/core/ports/repositories"
)

type ListPatientsUseCase struct {
	repo repositories.PatientRepository
}

func NewListPatients(repo repositories.PatientRepository) *ListPatientsUseCase {
	return &ListPatientsUseCase{repo: repo}
}

func (uc *ListPatientsUseCase) Execute(
	ctx context.Context,
	limit, offset int,
) ([]*patient.Patient, error) {
	patients, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	outputs := make([]*patient.Patient, len(patients))
	for i, p := range patients {
		outputs[i] = &p
	}

	return outputs, nil
}
