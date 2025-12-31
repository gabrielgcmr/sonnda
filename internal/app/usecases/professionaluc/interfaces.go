package professionaluc

import (
	"context"

	"sonnda-api/internal/domain/entities/professional"

	"github.com/google/uuid"
)

type CreateProfessionalInput struct {
	UserID             uuid.UUID
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

type UpdateProfessionalInput struct {
	ProfileID          uuid.UUID
	RegistrationNumber *string
	RegistrationIssuer *string
	RegistrationState  *string
	Status             *professional.VerificationStatus
}

type CreateProfessionalUseCase interface {
	Execute(ctx context.Context, input CreateProfessionalInput) (*professional.Professional, error)
}

type GetProfessionalUseCase interface {
	Execute(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error)
}

type UpdateProfessionalUseCase interface {
	Execute(ctx context.Context, input UpdateProfessionalInput) (*professional.Professional, error)
}

type DeleteProfessionalUseCase interface {
	Execute(ctx context.Context, profileID uuid.UUID) error
}
