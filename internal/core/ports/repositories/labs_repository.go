package repositories

import (
	"context"
	"sonnda-api/internal/core/domain"

	"github.com/google/uuid"
)

type LabsRepository interface {
	Create(ctx context.Context, report *domain.LabReport) error
	FindByID(ctx context.Context, reportID uuid.UUID) (*domain.LabReport, error)
	FindByPatientID(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]domain.LabReport, error)
	ExistsBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
