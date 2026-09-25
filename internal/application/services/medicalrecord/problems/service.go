// internal/application/services/medicalrecord/problems/service.go
package problems

import (
	"context"

	problem "github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord/problem"
	"github.com/google/uuid"
)

type CreateInput struct {
	MedicalRecordID uuid.UUID
	Name            string
	Abbreviation    string
	BodySystem      string
	Description     string
	Other           string
}

type UpdateInput struct {
	Name         *string
	Abbreviation *string
	BodySystem   *string
	Description  *string
	Other        *string
}

type Service interface {
	Create(ctx context.Context, input CreateInput) (*problem.Problem, error)
	Get(ctx context.Context, id uuid.UUID) (*problem.Problem, error)
	ListByMedicalRecord(ctx context.Context, medicalRecordID uuid.UUID, limit, offset int) ([]problem.Problem, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*problem.Problem, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
