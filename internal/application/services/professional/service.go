// internal/application/services/professional/service.go
package professionalsvc

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (*professional.Professional, error)
	GetByID(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error)
	Delete(ctx context.Context, profileID uuid.UUID) error
}
