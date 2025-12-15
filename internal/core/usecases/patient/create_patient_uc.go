package patient

import (
	"context"
	"time"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/google/uuid"
)

// vindo do HTTP (NÃO tem AppUserID)
type CreatePatientRequest struct {
	CPF       string        `json:"cpf"`
	CNS       *string       `json:"cns,omitempty"`
	FullName  string        `json:"full_name"`
	BirthDate time.Time     `json:"birth_date"`
	Gender    domain.Gender `json:"gender"`
	Race      domain.Race   `json:"race"`
	Phone     *string       `json:"phone,omitempty"`
	AvatarURL string        `json:"avatar_url"`
}

// comando interno (tem AppUserID)
type CreatePatientCommand struct {
	AppUserID *uuid.UUID
	CreatePatientRequest
}

type CreatePatientUseCase struct {
	repo repositories.PatientRepository
}

func NewCreatePatient(repo repositories.PatientRepository) *CreatePatientUseCase {
	return &CreatePatientUseCase{repo: repo}
}

func (uc *CreatePatientUseCase) Execute(
	ctx context.Context,
	cmd CreatePatientCommand,
) (*domain.Patient, error) {

	// 1. Verifica duplicidade
	existing, err := uc.repo.FindByCPF(ctx, cmd.CPF)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrCPFAlreadyExists
	}

	// 2. Cria entidade de domínio
	patient, err := domain.NewPatient(
		cmd.AppUserID,
		cmd.CPF,
		cmd.CNS,
		cmd.FullName,
		cmd.BirthDate,
		cmd.Gender,
		cmd.Race,
		cmd.Phone,
		cmd.AvatarURL,
	)
	if err != nil {
		return nil, err
	}

	// 4. Persiste
	if err := uc.repo.Create(ctx, patient); err != nil {
		return nil, err
	}

	// 5. Retorna output
	return patient, nil
}
