package usersvc

import (
	"time"

	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type ProfessionalRegisterInput struct {
	Kind               professional.Kind
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

type UserRegisterInput struct {
	Provider    string
	Subject     string
	Email       string
	AccountType user.AccountType
	FullName    string
	BirthDate   time.Time
	CPF         string
	Phone       string

	Professional *ProfessionalRegisterInput
}

type UserUpdateInput struct {
	UserID    uuid.UUID
	FullName  *string
	BirthDate *time.Time
	CPF       *string
	Phone     *string
}
