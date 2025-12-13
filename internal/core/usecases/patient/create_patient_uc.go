package patient

import (
	"context"
	"fmt"
	"time"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/google/uuid"
)

type CreatePatientInput struct {
	AppUserID *uuid.UUID    `json:"app_user_id,omitempty"`
	CPF       string        `json:"cpf"`
	CNS       *string       `json:"cns,omitempty"`
	FullName  string        `json:"full_name"`
	BirthDate time.Time     `json:"birth_date"` // já parseado na camada acima (handler)
	Gender    domain.Gender `json:"gender"`
	Race      domain.Race   `json:"race"`
	Phone     *string       `json:"phone,omitempty"`
	AvatarURL string        `json:"avatar_url"`
}

type CreatePatientUseCase struct {
	repo repositories.PatientRepository
}

func NewCreatePatient(repo repositories.PatientRepository) *CreatePatientUseCase {
	return &CreatePatientUseCase{repo: repo}
}

func (uc *CreatePatientUseCase) Execute(
	ctx context.Context,
	input CreatePatientInput,
) (*domain.Patient, error) {

	// 1. Verifica duplicidade
	existing, err := uc.repo.FindByCPF(ctx, input.CPF)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrCPFAlreadyExists
	}

	// 2. Cria entidade de domínio
	patient, err := domain.NewPatient(
		input.AppUserID,
		input.CPF,
		input.CNS,
		input.FullName,
		input.BirthDate,
		input.Gender,
		input.Race,
		input.Phone,
		input.AvatarURL,
	)
	if err != nil {
		return nil, err
	}

	fmt.Println("Created patient entity:", patient)

	// 4. Persiste
	if err := uc.repo.Create(ctx, patient); err != nil {
		return nil, err
	}

	// 5. Retorna output
	return patient, nil
}
