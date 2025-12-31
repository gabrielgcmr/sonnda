package user

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/entities/user"
	"sonnda-api/internal/domain/ports/repositories"
	"sonnda-api/internal/http/api/handlers/common"
	"sonnda-api/internal/http/middleware"

	applog "sonnda-api/internal/app/observability"
)

type UpdateUserRequest struct {
	FullName  *string `json:"full_name,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
	CPF       *string `json:"cpf,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}
type UserHandler struct {
	svc               usersvc.Service
	patientAccessRepo repositories.PatientAccessRepository
}

func NewUserHandler(
	svc usersvc.Service,
	patientAccessRepo repositories.PatientAccessRepository,
) *UserHandler {
	return &UserHandler{
		svc:               svc,
		patientAccessRepo: patientAccessRepo,
	}
}

type RegisterRequest struct {
	Email        string                   `json:"email"`
	FullName     string                   `json:"full_name" binding:"required"`
	BirthDate    string                   `json:"birth_date" binding:"required,datetime=2006-01-02"` // Gin já valida data!
	CPF          string                   `json:"cpf" binding:"required"`
	Phone        string                   `json:"phone" binding:"required"`
	Role         string                   `json:"role" binding:"required,oneof=caregiver professional"`
	Professional *ProfessionalRequestData `json:"professional" binding:"required_if=Role professional"` // Magia do Gin
}
type ProfessionalRequestData struct {
	RegistrationNumber string  `json:"registration_number"`
	RegistrationIssuer string  `json:"registration_issuer"`
	RegistrationState  *string `json:"registration_state,omitempty"`
}

func (h *UserHandler) Register(c *gin.Context) {
	// 1. Auth (Infra)
	identity, ok := middleware.GetIdentity(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, "missing_identity", nil)
		return
	}
	// 2. Bind & Validate Formato (Infra)
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(
			c,
			http.StatusBadRequest,
			"invalid_input",
			err,
		)
		return
	}
	// 3. Normalização de Email (Regra de Interface)
	email := req.Email
	if identity.Email != "" {
		email = identity.Email // Token tem prioridade
	}

	// 4. Dispatcher (Decisão de Roteamento)
	role := user.Role(strings.ToLower(req.Role))

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_birth_date", err)
		return
	}

	input := usersvc.RegisterInput{
		Provider:  identity.Provider,
		Subject:   identity.Subject,
		Email:     email,
		FullName:  req.FullName,
		Role:      role,
		BirthDate: birthDate,
		CPF:       req.CPF,
		Phone:     req.Phone,
	}

	if role == user.RoleProfessional {
		// Gin já garantiu required_if, mas mantemos safe-check.
		if req.Professional == nil {
			common.RespondError(c, http.StatusBadRequest, "professional_registration_required", errors.New("missing professional"))
			return
		}

		input.Professional = &usersvc.ProfessionalRegistrationInput{
			RegistrationNumber: req.Professional.RegistrationNumber,
			RegistrationIssuer: req.Professional.RegistrationIssuer,
			RegistrationState:  req.Professional.RegistrationState,
		}
	}

	if h.svc == nil {
		common.RespondError(
			c,
			http.StatusInternalServerError,
			"register_user_not_configured",
			errors.New("user service not configured"),
		)
		return
	}

	created, err := h.svc.Register(c.Request.Context(), input)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	c.JSON(http.StatusOK, currentUser)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("user_update")

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_input", err)
		return
	}

	input := usersvc.UpdateInput{
		UserID: currentUser.ID,
	}

	if req.FullName != nil {
		name := strings.TrimSpace(*req.FullName)
		input.FullName = &name
	}

	if req.BirthDate != nil {
		parsed, err := parseBirthDate(*req.BirthDate)
		if err != nil {
			common.RespondError(c, http.StatusBadRequest, "invalid_birth_date", err)
			return
		}
		input.BirthDate = &parsed
	}

	if req.CPF != nil {
		cpf := strings.TrimSpace(*req.CPF)
		input.CPF = &cpf
	}

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		input.Phone = &phone
	}

	if h.svc == nil {
		common.RespondError(
			c,
			http.StatusInternalServerError,
			"update_user_not_configured",
			errors.New("user service not configured"),
		)
		return
	}

	updated, err := h.svc.Update(c.Request.Context(), input)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	log.Info("user_updated")
	c.JSON(http.StatusOK, updated)
}

func parseBirthDate(raw string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, errors.New("birth_date is required")
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"02/01/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("invalid birth_date format")
}
