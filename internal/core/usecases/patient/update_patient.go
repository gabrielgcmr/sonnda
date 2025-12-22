package patient

import (
	"context"

	"sonnda-api/internal/core/domain/demographics"
	"sonnda-api/internal/core/domain/identity"
	"sonnda-api/internal/core/domain/patient"
	"sonnda-api/internal/core/ports/repositories"
	"sonnda-api/internal/core/ports/services"
)

type PatientChanges struct {
	FullName  *string              `json:"full_name,omitempty"`
	Phone     *string              `json:"phone,omitempty"`
	AvatarURL *string              `json:"avatar_url,omitempty"`
	Gender    *demographics.Gender `json:"gender,omitempty"`
	Race      *demographics.Race   `json:"race,omitempty"`
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
	currentUser *identity.User,
	cpf string,
	input PatientChanges,
) (*patient.Patient, error) {
	// 1) Busca paciente
	p, err := uc.repo.FindByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	//2) Verifica autorização
	if !uc.authorization.CanEditPatient(ctx, currentUser, p) {
		return nil, identity.ErrAuthorizationForbidden
	}

	//3) Aplica mudanças de dominio
	p.ApplyUpdate(
		input.FullName,
		input.Phone,
		input.AvatarURL,
		input.Gender,
		input.Race,
		input.CNS,
	)

	// 4) Persiste
	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (uc *UpdatePatientUseCase) ExecuteByID(
	ctx context.Context,
	currentUser *identity.User,
	id string,
	input PatientChanges,
) (*patient.Patient, error) {
	// 1) Busca paciente
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}

	//2) Verifica autorização
	if !uc.authorization.CanEditPatient(ctx, currentUser, p) {
		return nil, identity.ErrAuthorizationForbidden
	}

	//3) Aplica mudanças de dominio
	p.ApplyUpdate(
		input.FullName,
		input.Phone,
		input.AvatarURL,
		input.Gender,
		input.Race,
		input.CNS,
	)

	// 4) Persiste
	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
