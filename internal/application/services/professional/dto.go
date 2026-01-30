package professionalsvc

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/model/professional"

	"github.com/google/uuid"
)

type CreateInput struct {
	UserID             uuid.UUID
	Kind               professional.Kind
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}
