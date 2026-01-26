package repository

import (
	"context"

	"sonnda-api/internal/domain/model/patientaccess"

	"github.com/google/uuid"
)

// AccessiblePatient representa os dados mínimos de um paciente para listagem na UI
type AccessiblePatient struct {
	PatientID    uuid.UUID
	FullName     string
	AvatarURL    *string
	RelationType string
}

// PatientAccess armazena e consulta permissões por paciente para um app user.
type PatientAccess interface {
	// Lista mínima de pacientes acessíveis (para UI) com paginação
	// Retorna: lista de pacientes, total count, erro
	ListAccessiblePatientsByUser(ctx context.Context, granteeID uuid.UUID, limit, offset int) ([]AccessiblePatient, int64, error)

	// Cria ou atualiza um vínculo (reativa se estava revogado)
	Upsert(ctx context.Context, access *patientaccess.PatientAccess) error

	// Verifica se o usuário tem acesso ativo ao paciente
	HasActiveAccess(ctx context.Context, patientID, granteeID uuid.UUID) (bool, error)
}
