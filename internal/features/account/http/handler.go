// internal/features/account/http/handler.go
package accounthttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	authhttp "github.com/gabrielgcmr/sonnda/internal/features/auth/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	openapi "github.com/gabrielgcmr/sonnda/internal/api/openapi/generated"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type userService interface {
	Update(ctx context.Context, input account.UserUpdateInput) (*accountdomain.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	ListMyPatients(ctx context.Context, userID uuid.UUID, limit, offset int) (*account.MyPatientsOutput, error)
}

type Handler struct {
	onboarding account.Onboarding
	userSvc    userService
}

func NewHandler(
	onboarding account.Onboarding,
	userSvc userService,

) *Handler {
	return &Handler{
		onboarding: onboarding,
		userSvc:    userSvc,
	}
}

func (h *Handler) CreateUser(c *gin.Context) {
	identity, ok := authhttp.GetIdentity(c)
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

	input := account.RegisterInput{
		Issuer:      identity.Issuer,
		Subject:     identity.Subject,
		Email:       email,
		FullName:    req.FullName,
		AccountType: accountdomain.AccountTypeBasicCare,
		BirthDate:   birthDate,
		CPF:         req.Cpf,
		Phone:       req.Phone,
	}

	created, err := h.onboarding.Register(c.Request.Context(), input)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusCreated, userResponse(created))
}

func (h *Handler) GetUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	c.JSON(http.StatusOK, userResponse(currentUser))
}

func (h *Handler) UpdateUser(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)

	var req openapi.UpdateUserRequest
	if err := helpers.BindJSON(c, &req); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	input := account.UserUpdateInput{
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

	c.JSON(http.StatusOK, userResponse(updated))
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

	c.JSON(http.StatusOK, patientsResponse(result))
}
