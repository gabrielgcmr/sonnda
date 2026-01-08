// internal/app/services/patient/service_impl.go
package patientsvc

import (
	"context"
	"errors"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/app/services/authorization"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type service struct {
	core *coreService
	auth authorization.Authorizer
}

var _ Service = (*service)(nil)

func New(repo repository.Patient, auth authorization.Authorizer) Service {
	return &service{
		core: newCore(repo),
		auth: auth,
	}
}

func (s *service) Create(ctx context.Context, currentUser *user.User, input CreateInput) (*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionCreatePatient, nil); err != nil {
		return nil, err
	}

	return s.core.Create(ctx, input)
}

func (s *service) GetByID(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionReadPatient, &id); err != nil {
		return nil, err
	}

	p, err := s.core.GetByID(ctx, id)
	if err != nil {
		return nil, maskNotFoundAsForbidden(err)
	}
	return p, nil
}

func (s *service) UpdateByID(ctx context.Context, currentUser *user.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionUpdatePatient, &id); err != nil {
		return nil, err
	}

	p, err := s.core.UpdateByID(ctx, id, input)
	if err != nil {
		return nil, maskNotFoundAsForbidden(err)
	}
	return p, nil
}

func (s *service) SoftDeleteByID(ctx context.Context, currentUser *user.User, id uuid.UUID) error {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionSoftDeletePatient, &id); err != nil {
		return err
	}

	if err := s.core.SoftDeleteByID(ctx, id); err != nil {
		return maskNotFoundAsForbidden(err)
	}
	return nil
}

func (s *service) HardDeleteByID(ctx context.Context, currentUser *user.User, id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) List(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionListPatients, nil); err != nil {
		return nil, err
	}

	return s.core.List(ctx, limit, offset)
}

func (s *service) ListByName(ctx context.Context, currentUser *user.User, name string, limit, offset int) ([]*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionListPatients, nil); err != nil {
		return nil, err
	}

	return s.core.ListByName(ctx, name, limit, offset)
}

func maskNotFoundAsForbidden(err error) error {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil && appErr.Code == apperr.NOT_FOUND {
		return apperr.Forbidden("acesso negado")
	}
	return err
}
