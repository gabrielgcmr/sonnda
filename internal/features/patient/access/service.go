// internal/features/patient/access/service.go
package patientaccess

import (
	"context"
	"errors"
	"fmt"

	domainrepository "github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type Service interface {
	ListForAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) (*ListPatientsOutput, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListForAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) (*ListPatientsOutput, error) {
	if accountID == uuid.Nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}
	if s.repo == nil {
		return nil, apperr.Internal("erro inesperado", errors.New("patient access repository not configured"))
	}

	limit, offset = normalizePagination(limit, offset)
	patients, total, err := s.repo.ListAccessiblePatientsByUser(ctx, accountID, limit, offset)
	if err != nil {
		kind := apperr.INTERNAL_ERROR
		message := "erro inesperado"
		if errors.Is(err, domainrepository.ErrRepositoryFailure) {
			kind = apperr.INFRA_DATABASE_ERROR
			message = "falha técnica"
		}
		return nil, &apperr.AppError{
			Kind:    kind,
			Message: message,
			Cause:   fmt.Errorf("accessRepo.ListAccessiblePatientsByUser: %w", err),
		}
	}

	summaries := make([]PatientSummary, len(patients))
	for i, patient := range patients {
		summaries[i] = PatientSummary{
			ID:        patient.PatientID,
			FullName:  patient.FullName,
			AvatarURL: patient.AvatarURL,
		}
	}

	return &ListPatientsOutput{
		Patients: summaries,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

var _ Service = (*service)(nil)
