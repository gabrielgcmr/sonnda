// internal/features/patient/http/create_handler.go
package patienthttp

import (
	"fmt"
	"net/http"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	patientcreation "github.com/gabrielgcmr/sonnda/internal/application/usecase/patientcreation"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	openapi "github.com/gabrielgcmr/sonnda/internal/generated/openapi"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	applog "github.com/gabrielgcmr/sonnda/internal/kernel/observability"

	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type CreationHandler struct {
	creator patientcreation.UseCase
}

type createPatientRequest struct {
	CPF          string             `json:"cpf" binding:"required"`
	CNS          *string            `json:"cns,omitempty"`
	FullName     string             `json:"full_name" binding:"required"`
	BirthDate    openapi_types.Date `json:"birth_date" binding:"required"`
	Gender       string             `json:"gender" binding:"required"`
	Race         string             `json:"race" binding:"required"`
	Phone        *string            `json:"phone,omitempty"`
	AvatarURL    *string            `json:"avatar_url,omitempty"`
	RelationType string             `json:"relation_type" binding:"required"`
}

func NewCreationHandler(creator patientcreation.UseCase) *CreationHandler {
	return &CreationHandler{creator: creator}
}

func (h *CreationHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	applog.FromContext(ctx).Info("patient_create")

	creatorAccount, ok := helpers.GetCurrentUser(c)
	if !ok || creatorAccount == nil {
		presenter.ErrorResponder(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	var request createPatientRequest
	if err := helpers.BindJSON(c, &request); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	if request.BirthDate.Time.IsZero() {
		presenter.ErrorResponder(c, apperr.Validation(
			"data de nascimento é obrigatória",
			apperr.Violation{Field: "birth_date", Reason: "required"},
		))
		return
	}

	gender, err := parseGender(request.Gender)
	if err != nil {
		presenter.ErrorResponder(c, apperr.Validation(
			"gênero inválido",
			apperr.Violation{Field: "gender", Reason: "invalid"},
		))
		return
	}
	race, err := parseRace(request.Race)
	if err != nil {
		presenter.ErrorResponder(c, apperr.Validation(
			"raça inválida",
			apperr.Violation{Field: "race", Reason: "invalid"},
		))
		return
	}

	avatarURL := ""
	if request.AvatarURL != nil {
		avatarURL = *request.AvatarURL
	}

	patient, err := h.creator.Execute(ctx, creatorAccount.ID, patientcreation.Input{
		Profile: patientprofile.CreateInput{
			CPF:       request.CPF,
			CNS:       request.CNS,
			FullName:  request.FullName,
			BirthDate: request.BirthDate.Time,
			Gender:    gender,
			Race:      race,
			Phone:     request.Phone,
			AvatarURL: avatarURL,
		},
		Access: patientcreation.AccessInput{
			RelationType: request.RelationType,
		},
	})
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.Header("Location", "/v1/patients/"+patient.ID.String())
	c.JSON(http.StatusCreated, openapi.PatientCreatedResponse{Id: openapi_types.UUID(patient.ID)})
}

func parseGender(value string) (demographics.Gender, error) {
	gender, err := demographics.ParseGender(value)
	if err != nil {
		return "", fmt.Errorf("invalid gender value %q: %w", value, err)
	}
	return gender, nil
}

func parseRace(value string) (demographics.Race, error) {
	race, err := demographics.ParseRace(value)
	if err != nil {
		return "", fmt.Errorf("invalid race value %q: %w", value, err)
	}
	return race, nil
}
