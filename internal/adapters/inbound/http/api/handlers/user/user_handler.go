package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/api/handlers/common"
	httperrors "sonnda-api/internal/adapters/inbound/http/errors"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/app/apperr"
	registrationsvc "sonnda-api/internal/app/services/registration"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"
)

type Handler struct {
	regSvc  registrationsvc.Service
	userSvc usersvc.Service
}

func NewHandler(
	regSvc registrationsvc.Service,
	userSvc usersvc.Service,

) *Handler {
	return &Handler{
		regSvc:  regSvc,
		userSvc: userSvc,
	}
}

func (h *Handler) Register(c *gin.Context) {
	if h == nil || h.regSvc == nil {
		httperrors.WriteError(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	identity, ok := middleware.GetIdentity(c)
	if !ok {
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	var req RegisterRequest
	if err := httperrors.BindJSON(c, &req); err != nil {
		httperrors.WriteError(c, err)
		return
	}

	accountType := user.AccountType(req.AccountType).Normalize()

	birthDate, err := common.ParseBirthDate(req.BirthDate)
	if err != nil {
		httperrors.WriteError(c, apperr.Validation("data de nascimento inválida",
			apperr.Violation{
				Field:  "birth_date",
				Reason: "invalid_format",
			}))
		return
	}

	input := registrationsvc.RegisterInput{
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
		input.Professional = &registrationsvc.ProfessionalInput{
			Kind:               kind,
			RegistrationNumber: req.Professional.RegistrationNumber,
			RegistrationIssuer: req.Professional.RegistrationIssuer,
			RegistrationState:  req.Professional.RegistrationState,
		}
	}

	created, err := h.regSvc.Register(c.Request.Context(), input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) GetUser(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	c.JSON(http.StatusOK, currentUser)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	if h == nil || h.userSvc == nil {
		httperrors.WriteError(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	var req UpdateUserRequest
	if err := httperrors.BindJSON(c, &req); err != nil {
		httperrors.WriteError(c, err)
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
		parsed, _ := common.ParseBirthDate(*req.BirthDate)
		input.BirthDate = &parsed
	}

	updated, err := h.userSvc.Update(c.Request.Context(), input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) HardDeleteUser(c *gin.Context) {
	if h == nil || h.userSvc == nil {
		httperrors.WriteError(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	if err := h.userSvc.Delete(c.Request.Context(), currentUser.ID); err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListMyPatients(c *gin.Context) {
	panic("unimplemented")
}
