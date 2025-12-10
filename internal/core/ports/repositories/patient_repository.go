// core/ports/repositories/patient_repo.go
package repositories

import (
	"context"

	"sonnda-api/internal/core/domain"
)

type PatientRepository interface {
	// Operações CRUD básicas
	Create(ctx context.Context, patient *domain.Patient) error
	Update(ctx context.Context, patient *domain.Patient) error
	Delete(ctx context.Context, id string) error

	// Finders
	FindByCPF(ctx context.Context, cpf string) (*domain.Patient, error)
	FindByID(ctx context.Context, id string) (*domain.Patient, error)
	FindByUserID(ctx context.Context, userID string) (*domain.Patient, error)
	List(ctx context.Context, limit, offset int) ([]domain.Patient, error)
}
