// internal/features/account/onboarding_dto.go
package account

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
)

type ProfessionalInput struct {
	Kind               professional.Kind
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

type RegisterInput struct {
	Issuer      string
	Subject     string
	Email       string
	AccountType user.AccountType
	FullName    string
	BirthDate   time.Time
	CPF         string
	Phone       string

	Professional *ProfessionalInput
}
