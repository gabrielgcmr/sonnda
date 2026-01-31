// internal/domain/repository/labs.go
package repository

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"

	"github.com/google/uuid"
)

type Labs interface {
	// CRUD basico
	Create(ctx context.Context, report *labs.LabReport) error
	FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error)
	ExistsBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Listas
	ListLabs(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]labs.LabReport, error)
	ListItemsByPatientAndParameter(
		ctx context.Context,
		patientID uuid.UUID,
		parameterName string,
		limit, offset int,
	) ([]labs.LabResultItemTimeline, error)
}
