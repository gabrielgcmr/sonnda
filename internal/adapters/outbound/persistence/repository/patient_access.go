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

// Find implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) Find(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// FindActive implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) FindActive(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

var _ repository.PatientAccessRepository = (*PatientAccessRepository)(nil)

func NewPatientAccessRepository(client *db.Client) repository.PatientAccessRepository {
	return &PatientAccessRepository{
		client:  client,
		queries: patientaccesssqlc.New(client.Pool()),
	}
}

// ListByPatient implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// ListByUser implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// GetByUserAndPatient implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) GetByUserAndPatient(ctx context.Context, userID uuid.UUID, patientID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// Upsert implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	panic("unimplemented")
}

// Revoke implements [repository.PatientAccessRepository].
func (p *PatientAccessRepository) Revoke(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) error {
	panic("unimplemented")
}
