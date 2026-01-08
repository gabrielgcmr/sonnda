package repository

import (
	"context"

	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	patientaccesssqlc "sonnda-api/internal/adapters/outbound/persistence/sqlc/generated/patientaccess"
	"sonnda-api/internal/domain/model/patientaccess"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type PatientAccessRepository struct {
	client  *db.Client
	queries *patientaccesssqlc.Queries
}

// ListPatientAcessByUser implements [repository.PatientAccess].
func (p *PatientAccessRepository) ListPatientAcessByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// Find implements [repository.PatientAccess].
func (p *PatientAccessRepository) Find(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// FindActive implements [repository.PatientAccess].
func (p *PatientAccessRepository) FindActive(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

var _ repository.PatientAccess = (*PatientAccessRepository)(nil)

func NewPatientAccessRepository(client *db.Client) repository.PatientAccess {
	return &PatientAccessRepository{
		client:  client,
		queries: patientaccesssqlc.New(client.Pool()),
	}
}

// ListByPatient implements [repository.PatientAccess].
func (p *PatientAccessRepository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// ListByUser implements [repository.PatientAccess].
func (p *PatientAccessRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// GetByUserAndPatient implements [repository.PatientAccess].
func (p *PatientAccessRepository) GetByUserAndPatient(ctx context.Context, userID uuid.UUID, patientID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// Upsert implements [repository.PatientAccess].
func (p *PatientAccessRepository) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	panic("unimplemented")
}

// Revoke implements [repository.PatientAccess].
func (p *PatientAccessRepository) Revoke(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) error {
	panic("unimplemented")
}
