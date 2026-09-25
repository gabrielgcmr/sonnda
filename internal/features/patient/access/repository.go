// internal/features/patient/access/repository.go
package patientaccess

import (
	"context"

	accessdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/access/domain"

	"github.com/google/uuid"
)

// AccessiblePatient representa os dados mínimos de um paciente para listagem na UI
type AccessiblePatient struct {
	PatientID    uuid.UUID
	FullName     string
	AvatarURL    *string
	RelationType string
}

// Repository stores and queries account access to patients.
type Repository interface {
	// Lista mínima de pacientes acessíveis (para UI) com paginação
	// Retorna: lista de pacientes, total count, erro
	ListAccessiblePatientsByUser(ctx context.Context, granteeID uuid.UUID, limit, offset int) ([]AccessiblePatient, int64, error)

	// Cria ou atualiza um vínculo (reativa se estava revogado)
	Upsert(ctx context.Context, access *accessdomain.PatientAccess) error

	// Verifica se o usuário tem acesso ativo ao paciente
	HasActiveAccess(ctx context.Context, patientID, granteeID uuid.UUID) (bool, error)
}
