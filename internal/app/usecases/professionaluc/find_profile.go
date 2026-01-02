package professionaluc

import (
	"context"

	"sonnda-api/internal/domain/model/user/professional"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

type FindProfile struct {
	repo repositories.ProfessionalRepository
}

func NewFindProfile(repo repositories.ProfessionalRepository) *FindProfile {
	return &FindProfile{repo: repo}
}

func (uc *FindProfile) Execute(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error) {
	return uc.repo.FindByID(ctx, profileID)
}
