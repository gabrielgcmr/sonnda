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

// MyPatientsOutput represents the paginated list of patients accessible by the user
type MyPatientsOutput struct {
	Patients []PatientSummary `json:"patients"`
	Total    int64            `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
}

// PatientSummary represents minimal patient data for listing
type PatientSummary struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	RelationType string    `json:"relation_type"`
}
