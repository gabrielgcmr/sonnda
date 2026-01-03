package professionaluc

import (
	"context"

	"sonnda-api/internal/app/ports/outbound/repositories"
	"sonnda-api/internal/domain/model/user/professional"

	"github.com/google/uuid"
)

type DeleteProfile struct {
	repo repositories.ProfessionalRepository
}

func NewDeleteProfile(repo repositories.ProfessionalRepository) *DeleteProfile {
	return &DeleteProfile{repo: repo}
}

func (uc *DeleteProfile) Execute(ctx context.Context, profileID uuid.UUID) error {
	existing, err := uc.repo.FindByID(ctx, profileID)
	if err != nil {
		return err
	}
	if existing == nil {
		return professional.ErrProfileNotFound
	}

	return uc.repo.Delete(ctx, profileID)
}
