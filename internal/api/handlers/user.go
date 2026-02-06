// internal/api/handlers/user.go
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	openapi "github.com/gabrielgcmr/sonnda/internal/api/openapi/generated"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	usersvc "github.com/gabrielgcmr/sonnda/internal/application/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/registration"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type userService interface {
	Update(ctx context.Context, input usersvc.UserUpdateInput) (*user.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	ListMyPatients(ctx context.Context, userID uuid.UUID, limit, offset int) (*usersvc.MyPatientsOutput, error)
}

type UserHandler struct {
	regUC   registrationuc.UseCase
	userSvc userService
}

func NewUserHandler(
	regUC registrationuc.UseCase,
	userSvc userService,

) *UserHandler {
	return &UserHandler{
		regUC:   regUC,
		userSvc: userSvc,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	identity, ok := helpers.GetIdentity(c)
	if !ok {
		presenter.ErrorResponder(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	var req openapi.CreateUserRequest
	if err := helpers.BindJSON(c, &req); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	if req.BirthDate.Time.IsZero() {
		presenter.ErrorResponder(c, apperr.Validation("data de nascimento é obrigatória",
			apperr.Violation{
				Field:  "birth_date",
				Reason: "required",
			}))
		return
	}
	birthDate := req.BirthDate.Time

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
		AccountType: user.AccountTypeBasicCare,
		BirthDate:   birthDate,
		CPF:         req.Cpf,
		Phone:       req.Phone,
	}

	created, err := h.regUC.Register(c.Request.Context(), input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	c.JSON(http.StatusOK, currentUser)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	var req openapi.UpdateUserRequest
	if err := helpers.BindJSON(c, &req); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	input := usersvc.UserUpdateInput{
		UserID: currentUser.ID,
		CPF:    req.Cpf,
		Phone:  req.Phone,
	}

	if req.FullName != nil {
		input.FullName = req.FullName
	}
	if req.BirthDate != nil {
		parsed := req.BirthDate.Time
		input.BirthDate = &parsed
	}

	updated, err := h.userSvc.Update(c.Request.Context(), input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *UserHandler) HardDeleteUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	if err := h.userSvc.Delete(c.Request.Context(), currentUser.ID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.Status(http.StatusNoContent)

}

func (h *UserHandler) ListMyPatients(c *gin.Context) {
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
