package patient

import (
	"context"
	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
)

type GetPatientUseCase struct {
	repo repositories.PatientRepository
}

func NewGetPatient(repo repositories.PatientRepository) *GetPatientUseCase {
	return &GetPatientUseCase{repo: repo}
}

func (uc *GetPatientUseCase) ExecuteByCPF(ctx context.Context, cpf string) (*domain.Patient, error) {
	patient, err := uc.repo.FindByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	return patient, nil
}

func (uc *GetPatientUseCase) ExecuteByID(ctx context.Context, id string) (*domain.Patient, error) {
	patient, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	return patient, nil
}

func (uc *GetPatientUseCase) ExecuteByUserID(ctx context.Context, userID string) (*domain.Patient, error) {
	patient, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	return patient, nil
}

func (uc *GetPatientUseCase) ExecuteForUser(
	ctx context.Context,
	user *domain.User,
) (*domain.Patient, error) {

	patient, err := uc.ExecuteByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// converte domain.Patient -> PatientOutput como você já faz
	return patient, nil
}
