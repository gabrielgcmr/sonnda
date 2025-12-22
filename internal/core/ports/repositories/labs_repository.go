package repositories

import (
	"context"
	"sonnda-api/internal/core/domain/medicalRecord/exam/lab"
)

type LabsRepository interface {
	Create(ctx context.Context, report *lab.LabReport) error
	FindByID(ctx context.Context, reportID string) (*lab.LabReport, error)
	FindByPatientID(ctx context.Context, patientID string, limit, offset int) ([]lab.LabReport, error)
	ExistsBySignature(ctx context.Context, patientID string, fingerprint string) (bool, error)
	Delete(ctx context.Context, id string) error
}
