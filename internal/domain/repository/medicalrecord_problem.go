// internal/domain/repository/medicalrecord_problem.go
package repository

import (
	"context"

	problem "github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord/problem"
	"github.com/google/uuid"
)

type MedicalRecordProblem interface {
	Create(ctx context.Context, problem *problem.Problem) error
	Update(ctx context.Context, problem *problem.Problem) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*problem.Problem, error)
	ListByMedicalRecordID(ctx context.Context, medicalRecordID uuid.UUID, limit, offset int) ([]problem.Problem, error)
}
