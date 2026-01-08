package usersvc

import (
	"time"

	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type UserCreateInput struct {
	Provider    string
	Subject     string
	Email       string
	AccountType user.AccountType
	FullName    string
	BirthDate   time.Time
	CPF         string
	Phone       string
}

type UserUpdateInput struct {
	UserID    uuid.UUID
	FullName  *string
	BirthDate *time.Time
	CPF       *string
	Phone     *string
}
