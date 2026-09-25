// internal/features/account/onboarding_dto.go
package account

import (
	"time"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
)

type RegisterInput struct {
	Issuer      string
	Subject     string
	Email       string
	AccountType accountdomain.AccountType
	FullName    string
	BirthDate   time.Time
	CPF         string
	Phone       string
}
