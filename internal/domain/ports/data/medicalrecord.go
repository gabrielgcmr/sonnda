package data

import (
	"context"

	"sonnda-api/internal/domain/model/medicalrecord"

	"github.com/google/uuid"
)

type MedicalRecordRepo interface {
	// CRUD basico
	Save(ctx context.Context, record *medicalrecord.MedicalRecord) error
	Update(ctx context.Context, record *medicalrecord.MedicalRecord) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Finders
	FindByID(ctx context.Context, id uuid.UUID) (*medicalrecord.MedicalRecord, error)
	FindByPatientID(ctx context.Context, patientID uuid.UUID) (*medicalrecord.MedicalRecord, error)
	// Timeline
	CreateEntry(ctx context.Context, entry *medicalrecord.Entry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	ListEntries(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]medicalrecord.Entry, error)
}
