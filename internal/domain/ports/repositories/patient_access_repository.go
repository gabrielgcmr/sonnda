package repositories

import (
	"context"

	"sonnda-api/internal/domain/model/patient/patientaccess"

	"github.com/google/uuid"
)

// PatientAccessRepository armazena e consulta permissões por paciente para um app user.
// Usado quando caregiver/professional acessa pacientes que não são "dele".
type PatientAccessRepository interface {
	// Retorna o vínculo (ativo ou revogado) se existir
	Find(ctx context.Context, patientID, userID uuid.UUID) (*patientaccess.PatientAccess, error)
	//Quais vinculos ativos um usuário tem?
	FindActive(ctx context.Context, patientID, userID uuid.UUID) ([]*patientaccess.PatientAccess, error)
	//Quais pacientes um usuário tem acesso?
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error)

	Upsert(ctx context.Context, access *patientaccess.PatientAccess) error
}
