package registration

import (
	"time"

	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"
)

type ProfessionalInput struct {
	Kind               professional.Kind
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

type RegisterInput struct {
	Provider    string
	Subject     string
	Email       string
	AccountType user.AccountType
	FullName    string
	BirthDate   time.Time
	CPF         string
	Phone       string

	Professional *ProfessionalInput
}
