package patient

import (
	"context"

	"sonnda-api/internal/domain/model/patient/patientaccess"
	"sonnda-api/internal/domain/ports/repositories"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	patientaccesssqlc "sonnda-api/internal/adapters/outbound/persistence/sqlc/generated/patientaccess"

	"github.com/google/uuid"
)

type PatientAccessRepository struct {
	client  *db.Client
	queries *patientaccesssqlc.Queries
}

// Find implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) Find(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// FindActive implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) FindActive(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

var _ repositories.PatientAccessRepository = (*PatientAccessRepository)(nil)

func NewPatientAccessRepository(client *db.Client) repositories.PatientAccessRepository {
	return &PatientAccessRepository{
		client:  client,
		queries: patientaccesssqlc.New(client.Pool()),
	}
}

// ListByPatient implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// ListByUser implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// GetByUserAndPatient implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) GetByUserAndPatient(ctx context.Context, userID uuid.UUID, patientID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// Upsert implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	panic("unimplemented")
}

// Revoke implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) Revoke(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) error {
	panic("unimplemented")
}
