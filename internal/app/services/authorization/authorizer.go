package authorization

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/app/apperr"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/rbac"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"

	"github.com/google/uuid"
)

type Authorizer interface {
	Require(ctx context.Context, actor *user.User, action rbac.Action, patientID *uuid.UUID) error
}

type Service struct {
	rbacPolicy        *rbac.RbacPolicy
	patientRepo       ports.PatientRepo
	patientAccessRepo ports.PatientAccessRepo
	profRepo          ports.ProfessionalRepo
}

func New(
	patientRepo ports.PatientRepo,
	patientAccessRepo ports.PatientAccessRepo,
	profRepo ports.ProfessionalRepo,
) *Service {
	return &Service{
		rbacPolicy:        rbac.NewRbacPolicy(),
		patientRepo:       patientRepo,
		patientAccessRepo: patientAccessRepo,
		profRepo:          profRepo,
	}
}

func (s *Service) Require(ctx context.Context, actor *user.User, action rbac.Action, patientID *uuid.UUID) error {
	if actor == nil {
		return &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		}
	}

	subject, err := s.buildSubject(ctx, actor, action)
	if err != nil {
		return err
	}

	if !s.rbacPolicy.CanPerform(subject, action) {
		return &apperr.AppError{
			Code:    apperr.ACTION_NOT_ALLOWED,
			Message: "ação não permitida",
		}
	}

	if !isPatientScoped(action) {
		return nil
	}

	if patientID == nil || *patientID == uuid.Nil {
		return &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   errors.New("patientID is required for patient-scoped action"),
		}
	}

	if err := s.requirePatientAccess(ctx, actor.ID, *patientID); err != nil {
		return err
	}

	return nil
}

func (s *Service) buildSubject(ctx context.Context, actor *user.User, action rbac.Action) (rbac.Subject, error) {
	subject := rbac.Subject{
		AccountType: actor.AccountType,
	}

	if actor.AccountType != user.AccountTypeProfessional {
		return subject, nil
	}

	if !requiresProfessionalKind(action) {
		return subject, nil
	}

	if s.profRepo == nil {
		return rbac.Subject{}, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   errors.New("professional repository not configured"),
		}
	}

	prof, err := s.profRepo.FindByUserID(ctx, actor.ID)
	if err != nil {
		return rbac.Subject{}, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("profRepo.FindByUserID: %w", err),
		}
	}
	if prof == nil {
		return rbac.Subject{}, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   fmt.Errorf("professional profile missing for user_id=%s", actor.ID),
		}
	}

	kind := prof.Kind.Normalize()
	subject.ProfessionalKind = &kind
	return subject, nil
}

func (s *Service) requirePatientAccess(ctx context.Context, actorID uuid.UUID, patientID uuid.UUID) error {
	if s.patientRepo == nil || s.patientAccessRepo == nil {
		return &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   errors.New("authorizer repositories not configured"),
		}
	}

	p, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("patientRepo.FindByID: %w", err),
		}
	}

	// Preferência do projeto: responder 403 (ACCESS_DENIED) mesmo quando paciente não existe,
	// para evitar vazar existência de recursos.
	if p == nil {
		return &apperr.AppError{
			Code:    apperr.ACCESS_DENIED,
			Message: "acesso negado",
		}
	}

	if p.OwnerUserID != nil && *p.OwnerUserID == actorID {
		return nil
	}

	hasAccess, err := s.patientAccessRepo.HasActiveAccess(ctx, patientID, actorID)
	if err != nil {
		return &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("patientAccessRepo.HasActiveAccess: %w", err),
		}
	}
	if hasAccess {
		return nil
	}

	return &apperr.AppError{
		Code:    apperr.ACCESS_DENIED,
		Message: "acesso negado",
	}
}

func isPatientScoped(action rbac.Action) bool {
	switch action {
	case rbac.ActionReadPatient,
		rbac.ActionUpdatePatient,
		rbac.ActionSoftDeletePatient,
		rbac.ActionRecordMeasurement,
		rbac.ActionWriteClinicalNote,
		rbac.ActionReadLabs,
		rbac.ActionUploadLabs,
		rbac.ActionReadPrescriptions,
		rbac.ActionWritePrescriptions:
		return true
	default:
		return false
	}
}

func requiresProfessionalKind(action rbac.Action) bool {
	switch action {
	case rbac.ActionWritePrescriptions:
		return true
	default:
		return false
	}
}

var _ Authorizer = (*Service)(nil)
