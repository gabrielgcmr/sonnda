// internal/features/patient/profile/dto.go
package patientprofile

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
	accessdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/access/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	UserID       *uuid.UUID
	CPF          string
	CNS          *string
	FullName     string
	BirthDate    time.Time
	Gender       demographics.Gender
	Race         demographics.Race
	Phone        *string
	AvatarURL    string
	RelationType *accessdomain.RelationshipType
}

type UpdateInput struct {
	FullName  *string
	Phone     *string
	AvatarURL *string
	Gender    *demographics.Gender
	Race      *demographics.Race
	CNS       *string
}
