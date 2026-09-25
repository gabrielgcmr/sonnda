// internal/features/patient/profile/repository.go
package patientprofile

import (
	"context"
	"errors"

	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"

	"github.com/google/uuid"
)

var (
	ErrPatientAlreadyExists = errors.New("patient already exists")
	ErrPatientNotFound      = errors.New("patient not found")
)

type Repository interface {
	Create(ctx context.Context, patient *profiledomain.Patient) error
	Update(ctx context.Context, patient *profiledomain.Patient) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByCPF(ctx context.Context, cpf string) (*profiledomain.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*profiledomain.Patient, error)
	FindByName(ctx context.Context, name string) ([]profiledomain.Patient, error)
	List(ctx context.Context, limit, offset int) ([]profiledomain.Patient, error)
	SearchByName(ctx context.Context, name string, limit, offset int) ([]profiledomain.Patient, error)
}
