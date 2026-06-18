// internal/api/handlers/patient.go
package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	openapi "github.com/gabrielgcmr/sonnda/internal/api/openapi/generated"
	patientsvc "github.com/gabrielgcmr/sonnda/internal/application/services/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patientaccess"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	applog "github.com/gabrielgcmr/sonnda/internal/kernel/observability"

	openapi_types "github.com/oapi-codegen/runtime/types"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
)

type patientService interface {
	Create(ctx context.Context, currentUser *user.User, input patientsvc.CreateInput) (*patient.Patient, error)
	Get(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error)
	Update(ctx context.Context, currentUser *user.User, id uuid.UUID, input patientsvc.UpdateInput) (*patient.Patient, error)
	HardDelete(ctx context.Context, currentUser *user.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error)
}

type PatientHandler struct {
	svc patientService
}

type createPatientRequest struct {
	Cpf          string             `json:"cpf" binding:"required"`
	Cns          *string            `json:"cns,omitempty"`
	FullName     string             `json:"full_name" binding:"required"`
	BirthDate    openapi_types.Date `json:"birth_date" binding:"required"`
	Gender       string             `json:"gender" binding:"required"`
	Race         string             `json:"race" binding:"required"`
	Phone        *string            `json:"phone,omitempty"`
	AvatarUrl    *string            `json:"avatar_url,omitempty"`
	RelationType *string            `json:"relation_type,omitempty"`
}

func NewPatientHandler(svc patientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

func (h *PatientHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	log := applog.FromContext(ctx)
	log.Info("patient_create")

	user, ok := helpers.GetCurrentUser(c)
	if !ok || user == nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}

	var req createPatientRequest
	// 1. Bind do request
	if err := helpers.BindJSON(c, &req); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	// 3. Parsing / normalização de fronteira
	if req.BirthDate.Time.IsZero() {
		presenter.ErrorResponder(c, apperr.Validation("data de nascimento é obrigatória",
			apperr.Violation{Field: "birth_date", Reason: "required"}))
		return
	}
	birthDate := req.BirthDate.Time

	gender, err := ParseGender(string(req.Gender))
	if err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "gênero inválido",
			Cause:   err,
		})
		return
	}

	race, err := ParseRace(string(req.Race))
	if err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "raça inválida",
			Cause:   err,
		})
		return
	}

	avatarURL := ""
	if req.AvatarUrl != nil {
		avatarURL = *req.AvatarUrl
	}

	var ownerUserID *uuid.UUID
	var relationType *patientaccess.RelationshipType
	if req.RelationType != nil && strings.TrimSpace(*req.RelationType) != "" {
		rt := patientaccess.RelationshipType(strings.TrimSpace(*req.RelationType))
		if !rt.IsValid() {
			presenter.ErrorResponder(c, &apperr.AppError{
				Kind:    apperr.VALIDATION_FAILED,
				Message: "vÃ­nculo com paciente invÃ¡lido",
				Cause:   patientaccess.ErrInvalidRelationshipType,
			})
			return
		}
		relationType = &rt
		if rt == patientaccess.RelationshipTypeSelf {
			ownerUserID = &user.ID
		}
	}

	// 4. Montagem do input da aplicação
	input := patientsvc.CreateInput{
		UserID:       ownerUserID,
		CPF:          req.Cpf,
		CNS:          req.Cns,
		FullName:     req.FullName,
		BirthDate:    birthDate,
		Gender:       gender,
		Race:         race,
		Phone:        req.Phone,
		AvatarURL:    avatarURL,
		RelationType: relationType,
	}

	// 5. Execução do use case
	p, err := h.svc.Create(ctx, user, input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.Header("Location", "/v1/patients/"+p.ID.String())
	c.JSON(http.StatusCreated, openapi.PatientCreatedResponse{
		Id: openapi_types.UUID(p.ID),
	})
}

func (h *PatientHandler) GetPatient(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	id := c.Param("id")
	if id == "" {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return
	}

	p, err := h.svc.Get(c.Request.Context(), currentUser, parsedID)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	id := c.Param("id")
	if id == "" {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return
	}

	var input patientsvc.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "payload inválido",
			Cause:   err,
		})
		return
	}

	p, err := h.svc.Update(c.Request.Context(), currentUser, parsedID, input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) ListPatients(c *gin.Context) {
	if h == nil || h.svc == nil {
		presenter.ErrorResponder(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	currentUser := helpers.MustGetCurrentUser(c)

	list, err := h.svc.ListMyPatients(c.Request.Context(), currentUser, 100, 0)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *PatientHandler) HardDeletePatient(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	id := c.Param("id")
	if id == "" {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.VALIDATION_FAILED,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, parseErr := uuid.Parse(id)
	if parseErr != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.INVALID_FIELD_FORMAT,
			Message: "patient_id inválido",
			Cause:   parseErr,
		})
		return
	}

	if err := h.svc.HardDelete(c.Request.Context(), currentUser, parsedID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.Status(http.StatusNoContent)

}
