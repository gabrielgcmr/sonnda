package usersvc

import (
	"time"

	"sonnda-api/internal/domain/model/rbac"

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
	Role      rbac.Role
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
