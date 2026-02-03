// internal/api/handlers/user/user_handler.go
package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	usersvc "github.com/gabrielgcmr/sonnda/internal/application/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/registration"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type Handler struct {
	regUC   registrationuc.UseCase
	userSvc usersvc.Service
}

func NewHandler(
	regUC registrationuc.UseCase,
	userSvc usersvc.Service,

) *Handler {
	return &Handler{
		regUC:   regUC,
		userSvc: userSvc,
	}
}

func (h *Handler) Register(c *gin.Context) {
	if h == nil || h.regUC == nil {
		presenter.ErrorResponder(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	identity, ok := helpers.GetIdentity(c)
	if !ok {
		presenter.ErrorResponder(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	var req RegisterRequest
	if err := helpers.BindJSON(c, &req); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	accountType := user.AccountType(req.AccountType).Normalize()

	birthDate, err := handlers.ParseBirthDate(req.BirthDate)
	if err != nil {
		presenter.ErrorResponder(c, apperr.Validation("data de nascimento inválida",
			apperr.Violation{
				Field:  "birth_date",
				Reason: "invalid_format",
			}))
		return
	}

	if identity.Email == nil || strings.TrimSpace(*identity.Email) == "" {
		presenter.ErrorResponder(c, apperr.Validation("email é obrigatório"))
		return
	}
	email := strings.TrimSpace(*identity.Email)

	input := registrationuc.RegisterInput{
		Issuer:      identity.Issuer,
		Subject:     identity.Subject,
		Email:       email,
		FullName:    req.FullName,
		AccountType: accountType,
		BirthDate:   birthDate,
		CPF:         req.CPF,
		Phone:       req.Phone,
	}

	if accountType == user.AccountTypeProfessional {
		kind := professional.Kind(req.Professional.Kind).Normalize()
		input.Professional = &registrationuc.ProfessionalInput{
			Kind:               kind,
			RegistrationNumber: req.Professional.RegistrationNumber,
			RegistrationIssuer: req.Professional.RegistrationIssuer,
			RegistrationState:  req.Professional.RegistrationState,
		}
	}

	created, err := h.regUC.Register(c.Request.Context(), input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) GetUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	c.JSON(http.StatusOK, currentUser)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	var req UpdateUserRequest
	if err := helpers.BindJSON(c, &req); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	input := usersvc.UserUpdateInput{
		UserID: currentUser.ID,
		CPF:    req.CPF,
		Phone:  req.Phone,
	}

	if req.FullName != nil {
		input.FullName = req.FullName
	}
	if req.BirthDate != nil {
		parsed, _ := handlers.ParseBirthDate(*req.BirthDate)
		input.BirthDate = &parsed
	}

	updated, err := h.userSvc.Update(c.Request.Context(), input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) HardDeleteUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	if err := h.userSvc.Delete(c.Request.Context(), currentUser.ID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.Status(http.StatusNoContent)

}

func (h *Handler) ListMyPatients(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	// Parse query params
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	result, err := h.userSvc.ListMyPatients(c.Request.Context(), currentUser.ID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
