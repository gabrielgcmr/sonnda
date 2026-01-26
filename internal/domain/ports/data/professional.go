package data

import (
	"context"

	"sonnda-api/internal/domain/model/professional"

	"github.com/google/uuid"
)

type ProfessionalRepo interface {
	// CRUD basico
	Create(ctx context.Context, profile *professional.Professional) error
	Update(ctx context.Context, profile *professional.Professional) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Buscas por atributos do profissional
	FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error)
	FindByRegistration(ctx context.Context, registrationNumber, registrationIssuer string) (*professional.Professional, error)
	FindByName(ctx context.Context, name string, limit, offset int) ([]*professional.Professional, error)
}
