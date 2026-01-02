package repositories

import (
	"context"

	"sonnda-api/internal/domain/entities/patientaccess"

	"github.com/google/uuid"
)

// PatientAccessRepository armazena e consulta permissões por paciente para um app user.
// Usado quando caregiver/professional acessa pacientes que não são "dele".
type PatientAccessRepository interface {
	// Retorna o vínculo (preferível para checar permissão).
	GetByUserAndPatient(ctx context.Context, userID, patientID uuid.UUID) (*patientaccess.PatientAccess, error)
	//Quais vinculos ativos um usuário tem?
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*patientaccess.PatientAccess, error)
	//Quais pacientes um usuário tem acesso?
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error)

	Upsert(ctx context.Context, access *patientaccess.PatientAccess) error
	Revoke(ctx context.Context, patientID, userID uuid.UUID) error
}
