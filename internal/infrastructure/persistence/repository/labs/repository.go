package labs

import (
	"context"

	"sonnda-api/internal/app/interfaces/repositories"
	"sonnda-api/internal/domain/model/medicalrecord/labs"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	labsqlc "sonnda-api/internal/infrastructure/persistence/sqlc/generated/lab"

	"github.com/google/uuid"
)

type LabsRepository struct {
	client  *db.Client
	queries *labsqlc.Queries
}

var _ repositories.LabRepository = (*LabsRepository)(nil)

func NewLabsRepository(client *db.Client) repositories.LabRepository {
	return &LabsRepository{
		client:  client,
		queries: labsqlc.New(client.Pool()),
	}
}

// Create implements [repositories.LabRepository].
func (l *LabsRepository) Create(ctx context.Context, report *labs.LabReport) error {
	panic("unimplemented")
}

// Delete implements [repositories.LabRepository].
func (l *LabsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// ExistsBySignature implements [repositories.LabRepository].
func (l *LabsRepository) ExistsBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (bool, error) {
	panic("unimplemented")
}

// FindByID implements [repositories.LabRepository].
func (l *LabsRepository) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	panic("unimplemented")
}

// ListItemsByPatientAndParameter implements [repositories.LabRepository].
func (l *LabsRepository) ListItemsByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit int, offset int) ([]labs.LabResultItemTimeline, error) {
	panic("unimplemented")
}

// ListLabs implements [repositories.LabRepository].
func (l *LabsRepository) ListLabs(ctx context.Context, patientID uuid.UUID, limit int, offset int) ([]labs.LabReport, error) {
	panic("unimplemented")
}
