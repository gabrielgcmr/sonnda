package professionalsvc

import (
	"context"

	"sonnda-api/internal/domain/model/professional"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (*professional.Professional, error)
	GetByID(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error)
	Delete(ctx context.Context, profileID uuid.UUID) error
}
