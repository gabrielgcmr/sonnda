// internal/features/patient/profile/repository.go
package patientprofile

import (
	"context"
	"errors"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patientaccess"

	"github.com/google/uuid"
)

var (
	ErrPatientAlreadyExists = errors.New("patient already exists")
	ErrPatientNotFound      = errors.New("patient not found")
)

type Repository interface {
	Create(ctx context.Context, patient *patient.Patient) error
	CreateWithAccess(ctx context.Context, patient *patient.Patient, access *patientaccess.PatientAccess) error
	Update(ctx context.Context, patient *patient.Patient) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error)
	FindByName(ctx context.Context, name string) ([]patient.Patient, error)
	List(ctx context.Context, limit, offset int) ([]patient.Patient, error)
	SearchByName(ctx context.Context, name string, limit, offset int) ([]patient.Patient, error)
}
