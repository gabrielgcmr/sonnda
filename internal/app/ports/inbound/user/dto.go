package userport

import (
	"time"

	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"

	"github.com/google/uuid"
)

type ProfessionalRegistrationInput struct {
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

	Professional *ProfessionalRegistrationInput
}

type UpdateInput struct {
	UserID    uuid.UUID
	FullName  *string
	BirthDate *time.Time
	CPF       *string
	Phone     *string
}
