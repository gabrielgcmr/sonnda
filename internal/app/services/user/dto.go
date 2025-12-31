package usersvc

import (
	"time"

	"sonnda-api/internal/domain/entities/user"

	"github.com/google/uuid"
)

type ProfessionalRegistrationInput struct {
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

type RegisterInput struct {
	Provider  string
	Subject   string
	Email     string
	Role      user.Role
	FullName  string
	BirthDate time.Time
	CPF       string
	Phone     string

	Professional *ProfessionalRegistrationInput
}

type UpdateInput struct {
	UserID    uuid.UUID
	FullName  *string
	BirthDate *time.Time
	CPF       *string
	Phone     *string
}
