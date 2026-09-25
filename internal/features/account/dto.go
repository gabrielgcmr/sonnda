// internal/features/account/dto.go
package account

import (
	"time"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"

	"github.com/google/uuid"
)

type UserCreateInput struct {
	Issuer      string
	Subject     string
	Email       string
	AccountType accountdomain.AccountType
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
