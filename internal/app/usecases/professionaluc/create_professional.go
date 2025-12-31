package professionaluc

import (
	"context"

	"sonnda-api/internal/domain/entities/professional"
	"sonnda-api/internal/domain/ports/repositories"
)

// Garante que implementa a interface
var _ CreateProfessionalUseCase = (*CreateProfessional)(nil)

// Renomeado de CreateProfile
type CreateProfessional struct {
	repo repositories.ProfessionalRepository
}

func NewCreateProfessional(repo repositories.ProfessionalRepository) *CreateProfessional {
	return &CreateProfessional{
		repo: repo,
	}
}

func (uc *CreateProfessional) Execute(ctx context.Context, input CreateProfessionalInput) (*professional.Professional, error) {
	// 1. Converter Input (App) para Params (Domain)
	// Como renomeamos tudo, fica muito mais legível: "Professional Input" cria "Professional Params"
	params := professional.NewProfessionalParams{
		UserID:             input.UserID,
		RegistrationNumber: input.RegistrationNumber,
		RegistrationIssuer: input.RegistrationIssuer,
		RegistrationState:  input.RegistrationState,
	}

	// 2. Criar Entidade
	prof, err := professional.NewProfessional(params)
	if err != nil {
		return nil, err
	}

	// 3. Persistir
	if err := uc.repo.Create(ctx, prof); err != nil {
		return nil, err
	}

	return prof, nil
}
