package repositories

import (
	"context"

	"sonnda-api/internal/domain/entities/patientaccess"

	"github.com/google/uuid"
)

// PatientAccessRepository armazena e consulta permissões por paciente para um app user.
// Usado quando caregiver/professional acessa pacientes que não são "dele".
type PatientAccessRepository interface {
	// Upsert cria ou atualiza o vínculo (ex.: mudar role).
	// Recomendação: manter UNIQUE(patient_id, user_id) no banco.
	Upsert(ctx context.Context, access *patientaccess.PatientAccess) error

	// Find retorna o vínculo específico (nil, nil se não existir).
	Find(ctx context.Context, patientID, userID uuid.UUID) (*patientaccess.PatientAccess, error)
	// ListByPatient lista todos os membros (caregivers/professionals) com acesso ao paciente.
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error)

	// ListByUser lista todos os pacientes que esse app user pode acessar.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*patientaccess.PatientAccess, error)

	// Revoke remove o vínculo (revoga acesso).
	Revoke(ctx context.Context, patientID, userID uuid.UUID) error

	// HasPermission é um atalho para autorização.
	// Pode ser implementado via query eficiente (sem carregar listas) ou via Find + domain.HasPermission.
	HasPermission(ctx context.Context, patientID, userID uuid.UUID, perm patientaccess.Permission) (bool, error)
}
