package patient

import (
	"context"

	"sonnda-api/internal/domain/entities/patientaccess"
	"sonnda-api/internal/domain/ports/repositories"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	patientaccesssqlc "sonnda-api/internal/infrastructure/persistence/sqlc/generated/patientaccess"

	"github.com/google/uuid"
)

type PatientAccessRepository struct {
	client  *db.Client
	queries *patientaccesssqlc.Queries
}

var _ repositories.PatientAccessRepository = (*PatientAccessRepository)(nil)

func NewPatientAccessRepository(client *db.Client) repositories.PatientAccessRepository {
	return &PatientAccessRepository{
		client:  client,
		queries: patientaccesssqlc.New(client.Pool()),
	}
}

// Find implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) Find(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) (*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// HasPermission implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) HasPermission(ctx context.Context, patientID uuid.UUID, userID uuid.UUID, perm patientaccess.Permission) (bool, error) {
	panic("unimplemented")
}

// ListByPatient implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) ListByPatient(ctx context.Context, patientID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// ListByUser implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*patientaccess.PatientAccess, error) {
	panic("unimplemented")
}

// Revoke implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) Revoke(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) error {
	panic("unimplemented")
}

// Upsert implements [repositories.PatientAccessRepository].
func (p *PatientAccessRepository) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	panic("unimplemented")
}
