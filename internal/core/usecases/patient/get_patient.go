package patient

import (
	"context"

	"sonnda-api/internal/core/domain/identity"
	"sonnda-api/internal/core/domain/patient"
	"sonnda-api/internal/core/ports/repositories"
)

type GetPatientUseCase struct {
	repo repositories.PatientRepository
}

func NewGetPatient(repo repositories.PatientRepository) *GetPatientUseCase {
	return &GetPatientUseCase{repo: repo}
}

func (uc *GetPatientUseCase) ExecuteByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	p, err := uc.repo.FindByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	return p, nil
}

func (uc *GetPatientUseCase) ExecuteByID(ctx context.Context, id string) (*patient.Patient, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	return p, nil
}

func (uc *GetPatientUseCase) ExecuteByUserID(ctx context.Context, userID string) (*patient.Patient, error) {
	p, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	return p, nil
}

func (uc *GetPatientUseCase) ExecuteForUser(
	ctx context.Context,
	user *identity.User,
) (*patient.Patient, error) {

	p, err := uc.ExecuteByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// converte domain.Patient -> PatientOutput como você já faz
	return p, nil
}
