package repository

import (
	"context"

	"sonnda-api/internal/domain/model/patientaccess"

	"github.com/google/uuid"
)

// PatientAccess armazena e consulta permissões por paciente para um app user.
// Usado quando caregiver/professional acessa pacientes que não são "dele".
type PatientAccess interface {
	// Lista todos os vínculos ativos
	ListPatientAcessByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*patientaccess.PatientAccess, error)

	// Retorna o vínculo (ativo ou revogado) se existir
	Find(ctx context.Context, patientID, userID uuid.UUID) (*patientaccess.PatientAccess, error)
	// Quais vínculos ativos um usuário tem?

	FindActive(ctx context.Context, patientID, userID uuid.UUID) ([]*patientaccess.PatientAccess, error)
	// Quais pacientes um usuário tem acesso?
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error)

	Upsert(ctx context.Context, access *patientaccess.PatientAccess) error
}
