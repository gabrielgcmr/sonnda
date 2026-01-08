package repository

import (
	"context"

	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	labsqlc "sonnda-api/internal/adapters/outbound/persistence/sqlc/generated/lab"
	"sonnda-api/internal/domain/model/medicalrecord/labs"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type LabsRepository struct {
	client  *db.Client
	queries *labsqlc.Queries
}

var _ repository.LabRepository = (*LabsRepository)(nil)

func NewLabsRepository(client *db.Client) repository.LabRepository {
	return &LabsRepository{
		client:  client,
		queries: labsqlc.New(client.Pool()),
	}
}

// Create implements [repository.LabRepository].
func (l *LabsRepository) Create(ctx context.Context, report *labs.LabReport) error {
	panic("unimplemented")
}

// Delete implements [repository.LabRepository].
func (l *LabsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// ExistsBySignature implements [repository.LabRepository].
func (l *LabsRepository) ExistsBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (bool, error) {
	panic("unimplemented")
}

// FindByID implements [repository.LabRepository].
func (l *LabsRepository) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	panic("unimplemented")
}

// ListItemsByPatientAndParameter implements [repository.LabRepository].
func (l *LabsRepository) ListItemsByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit int, offset int) ([]labs.LabResultItemTimeline, error) {
	panic("unimplemented")
}

// ListLabs implements [repository.LabRepository].
func (l *LabsRepository) ListLabs(ctx context.Context, patientID uuid.UUID, limit int, offset int) ([]labs.LabReport, error) {
	panic("unimplemented")
}
