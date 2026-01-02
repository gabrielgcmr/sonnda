package patientaccess

import (
	"time"

	"github.com/google/uuid"
)

type RelationshipType string //O relacionamento que o usuário tem com o paciente

const (
	RelationshipTypeCaregiver    RelationshipType = "caregiver"
	RelationshipTypeFamily       RelationshipType = "family"
	RelationshipTypeProfessional RelationshipType = "professional"
	RelationshipTypeSelf         RelationshipType = "self" // Quando o proprio paciente é também o usuário
)

//Não foi criado um status para o vínculo, pois ele se ele existe já é ativo ou não ser que tenha sido explicitamente revogado.

type PatientAccess struct {
	PatientID uuid.UUID
	UserID    uuid.UUID

	Type RelationshipType

	CreatedAt time.Time
	RevokedAt *time.Time //Se nulo, está ativo
	GrantedBy *uuid.UUID
}

// Lembrete Regras de dominio!
// Quem cria o paciente já nasce com um vinculo com ele.
// O paciente deve tem pelo menos um vinculo ativo.
