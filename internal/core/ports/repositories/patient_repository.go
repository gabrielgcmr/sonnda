// core/ports/repositories/patient_repo.go
package repositories

import (
	"context"

	"sonnda-api/internal/core/domain/patient"
)

type PatientRepository interface {
	// Operações CRUD básicas
	Create(ctx context.Context, patient *patient.Patient) error
	Update(ctx context.Context, patient *patient.Patient) error
	Delete(ctx context.Context, id string) error

	// Finders
	FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error)
	FindByID(ctx context.Context, id string) (*patient.Patient, error)
	FindByUserID(ctx context.Context, userID string) (*patient.Patient, error)
	List(ctx context.Context, limit, offset int) ([]patient.Patient, error)
}
