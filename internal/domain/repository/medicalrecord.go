// internal/domain/repository/medicalrecord.go
package repository

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord"

	"github.com/google/uuid"
)

type MedicalRecord interface {
	Create(ctx context.Context, record *medicalrecord.MedicalRecord) error
	Update(ctx context.Context, record *medicalrecord.MedicalRecord) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*medicalrecord.MedicalRecord, error)
	FindByPatientID(ctx context.Context, patientID uuid.UUID) (*medicalrecord.MedicalRecord, error)
}
