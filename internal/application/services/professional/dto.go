// internal/application/services/professional/dto.go
package professionalsvc

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"

	"github.com/google/uuid"
)

type CreateInput struct {
	UserID             uuid.UUID
	Kind               professional.Kind
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}
