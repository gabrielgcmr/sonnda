// internal/features/patient/profile/http/handler.go
package profilehttp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
)

type patientService interface {
	Get(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) (*profiledomain.Patient, error)
	Update(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID, input patientprofile.UpdateInput) (*profiledomain.Patient, error)
	HardDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *accountdomain.User, limit, offset int) ([]*profiledomain.Patient, error)
}

type Handler struct {
	svc patientService
}

func NewHandler(svc patientService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetPatient(c *gin.Context) {
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

func (h *Handler) UpdatePatient(c *gin.Context) {
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

	var input patientprofile.UpdateInput
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

func (h *Handler) ListPatients(c *gin.Context) {
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

func (h *Handler) HardDeletePatient(c *gin.Context) {
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
