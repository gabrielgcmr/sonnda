package repositories

import (
	"context"
	"sonnda-api/internal/core/domain"
)

type LabsRepository interface {
	Create(ctx context.Context, report *domain.LabReport) error
	FindByID(ctx context.Context, reportID string) (*domain.LabReport, error)
	FindByPatientID(ctx context.Context, patientID string, limit, offset int) ([]domain.LabReport, error)
	ExistsByPatientAndRawText(ctx context.Context, patientID, reportHash string) (bool, error)
	ExistsBySignature(ctx context.Context, patientID, reportHash string) (bool, error)
	Delete(ctx context.Context, id string) error
}
