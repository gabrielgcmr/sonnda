// internal/domain/repository/patient.go
package repository

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"

	"github.com/google/uuid"
)

type PatientRepo interface {
	// Operações CRUD básicas
	Create(ctx context.Context, patient *patient.Patient) error
	Update(ctx context.Context, patient *patient.Patient) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error

	// Finders
	FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error)
	FindByName(ctx context.Context, name string) ([]patient.Patient, error)
	// Listagem
	List(ctx context.Context, limit, offset int) ([]patient.Patient, error)
	//
	SearchByName(ctx context.Context, name string, limit, offset int) ([]patient.Patient, error)
}
