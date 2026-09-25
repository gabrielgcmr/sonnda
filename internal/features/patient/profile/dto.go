// internal/features/patient/profile/dto.go
package patientprofile

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
)

type CreateInput struct {
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
