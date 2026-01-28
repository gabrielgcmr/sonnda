package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/apierr"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/binder"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/httpctx"
	usersvc "github.com/gabrielgcmr/sonnda/internal/app/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/app/usecase/registration"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	"github.com/gabrielgcmr/sonnda/internal/shared/apperr"
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
		apierr.ErrorResponder(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	identity, ok := httpctx.GetIdentity(c)
	if !ok {
		apierr.ErrorResponder(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	var req RegisterRequest
	if err := binder.BindJSON(c, &req); err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	accountType := user.AccountType(req.AccountType).Normalize()

	birthDate, err := handlers.ParseBirthDate(req.BirthDate)
	if err != nil {
		apierr.ErrorResponder(c, apperr.Validation("data de nascimento inválida",
			apperr.Violation{
				Field:  "birth_date",
				Reason: "invalid_format",
			}))
		return
	}

	input := registrationuc.RegisterInput{
		Provider:    identity.Provider,
		Subject:     identity.Subject,
		Email:       identity.Email,
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
		apierr.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) GetUser(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)
	c.JSON(http.StatusOK, currentUser)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)

	var req UpdateUserRequest
	if err := binder.BindJSON(c, &req); err != nil {
		apierr.ErrorResponder(c, err)
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
		apierr.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) HardDeleteUser(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)

	if err := h.userSvc.Delete(c.Request.Context(), currentUser.ID); err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListMyPatients(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)

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
		apierr.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
