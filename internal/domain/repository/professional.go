// internal/domain/repository/professional.go
package repository

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"

	"github.com/google/uuid"
)

type Professional interface {
	Create(ctx context.Context, p *professional.Professional) error
	Update(ctx context.Context, p *professional.Professional) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error)
	FindByRegistration(ctx context.Context, registrationNumber, registrationIssuer string) (*professional.Professional, error)
	FindByName(ctx context.Context, name string, limit, offset int) ([]*professional.Professional, error)
}
