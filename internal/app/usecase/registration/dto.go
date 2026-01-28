package registration

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
)

type ProfessionalInput struct {
	Kind               professional.Kind
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

type RegisterInput struct {
	Email       string
	AccountType user.AccountType
	FullName    string
	BirthDate   time.Time
	CPF         string
	Phone       string

	Professional *ProfessionalInput
}
