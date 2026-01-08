// core/ports/repositories/patient_repo.go
package repository

import (
	"context"
	"time"

	"sonnda-api/internal/domain/model/patient"

	"github.com/google/uuid"
)

type Patient interface {
	// Operações CRUD básicas
	Create(ctx context.Context, patient *patient.Patient) error
	Update(ctx context.Context, patient *patient.Patient) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Finders
	FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error)
	List(ctx context.Context, limit, offset int) ([]patient.Patient, error)
	ListByName(ctx context.Context, name string, limit, offset int) ([]patient.Patient, error)
	ListByBirthDate(ctx context.Context, birthDate time.Time, limit, offset int) ([]patient.Patient, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID) ([]patient.Patient, error)
}
