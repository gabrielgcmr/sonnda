// internal/app/services/patient/dto.go
package patientsvc

import (
	"sonnda-api/internal/domain/model/demographics"
	"time"

	"github.com/google/uuid"
)

type CreateInput struct {
	UserID    *uuid.UUID
	CPF       string
	CNS       *string
	FullName  string
	BirthDate time.Time
	Gender    demographics.Gender
	Race      demographics.Race
	Phone     *string
	AvatarURL string
}

type UpdateInput struct {
	FullName  *string
	Phone     *string
	AvatarURL *string
	Gender    *demographics.Gender
	Race      *demographics.Race
	CNS       *string
}
