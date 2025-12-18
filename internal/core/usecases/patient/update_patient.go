package patient

import (
	"context"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
	"sonnda-api/internal/core/ports/services"

	"github.com/google/uuid"
)

type PatientChanges struct {
	FullName  *string        `json:"full_name,omitempty"`
	Phone     *string        `json:"phone,omitempty"`
	AvatarURL *string        `json:"avatar_url,omitempty"`
	Gender    *domain.Gender `json:"gender,omitempty"`
	Race      *domain.Race   `json:"race,omitempty"`
	// Se um dia você quiser permitir corrigir CNS:
	CNS *string `json:"cns,omitempty"`
}
type UpdatePatientUseCase struct {
	repo          repositories.PatientRepository
	authorization services.AuthorizationService
}

func NewUpdatePatient(
	repo repositories.PatientRepository,
	authorization services.AuthorizationService,
) *UpdatePatientUseCase {
	return &UpdatePatientUseCase{
		repo:          repo,
		authorization: authorization,
	}
}

func (uc *UpdatePatientUseCase) ExecuteByCPF(
	ctx context.Context,
	currentUser *domain.User,
	cpf string,
	input PatientChanges,
) (*domain.Patient, error) {
	// 1) Busca paciente
	patient, err := uc.repo.FindByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	//2) Verifica autorização
	if !uc.authorization.CanEditPatient(ctx, currentUser, patient) {
		return nil, domain.ErrForbidden
	}

	//3) Aplica mudanças de dominio
	patient.ApplyUpdate(
		input.FullName,
		input.Phone,
		input.AvatarURL,
		input.Gender,
		input.Race,
		input.CNS,
	)

	// 4) Persiste
	if err := uc.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	return patient, nil
}

func (uc *UpdatePatientUseCase) ExecuteByID(
	ctx context.Context,
	currentUser *domain.User,
	id uuid.UUID,
	input PatientChanges,
) (*domain.Patient, error) {
	// 1) Busca paciente
	patient, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, domain.ErrPatientNotFound
	}

	//2) Verifica autorização
	if !uc.authorization.CanEditPatient(ctx, currentUser, patient) {
		return nil, domain.ErrForbidden
	}

	//3) Aplica mudanças de dominio
	patient.ApplyUpdate(
		input.FullName,
		input.Phone,
		input.AvatarURL,
		input.Gender,
		input.Race,
		input.CNS,
	)

	// 4) Persiste
	if err := uc.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	return patient, nil
}
