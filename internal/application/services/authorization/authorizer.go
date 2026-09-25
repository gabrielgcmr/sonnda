// internal/application/services/authorization/authorizer.go
package authorization

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

// Authorizer checks access to a patient independently of account type or action.
type Authorizer interface {
	RequirePatientAccess(ctx context.Context, actor *accountdomain.User, patientID uuid.UUID) error
}

type Service struct {
	patientRepo       patientprofile.Repository
	patientAccessRepo repository.PatientAccessRepo
}

func New(patientRepo patientprofile.Repository, patientAccessRepo repository.PatientAccessRepo) *Service {
	return &Service{patientRepo: patientRepo, patientAccessRepo: patientAccessRepo}
}

func (s *Service) RequirePatientAccess(ctx context.Context, actor *accountdomain.User, patientID uuid.UUID) error {
	if actor == nil || actor.ID == uuid.Nil {
		return apperr.Unauthorized("autenticação necessária")
	}
	if patientID == uuid.Nil {
		return apperr.Internal("erro inesperado", errors.New("patientID is required"))
	}
	if s.patientRepo == nil || s.patientAccessRepo == nil {
		return apperr.Internal("erro inesperado", errors.New("patient access repositories not configured"))
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return &apperr.AppError{
			Kind:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("patientRepo.FindByID: %w", err),
		}
	}
	// Return the same error for a missing patient and a patient without access.
	if p == nil {
		return apperr.Forbidden("acesso negado")
	}
	if p.OwnerUserID != nil && *p.OwnerUserID == actor.ID {
		return nil
	}

	hasAccess, err := s.patientAccessRepo.HasActiveAccess(ctx, patientID, actor.ID)
	if err != nil {
		return &apperr.AppError{
			Kind:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("patientAccessRepo.HasActiveAccess: %w", err),
		}
	}
	if !hasAccess {
		return apperr.Forbidden("acesso negado")
	}
	return nil
}

var _ Authorizer = (*Service)(nil)
