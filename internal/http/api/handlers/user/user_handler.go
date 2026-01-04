package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/app/apperr"
	userport "sonnda-api/internal/app/ports/inbound/user"
	"sonnda-api/internal/app/ports/outbound/repositories"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"
	"sonnda-api/internal/http/api/handlers/common"
	httperrors "sonnda-api/internal/http/errors"
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
	svc               userport.UserService
	patientAccessRepo repositories.PatientAccessRepository
}

func NewUserHandler(
	svc userport.UserService,
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
	AccountType  string                   `json:"account_type" binding:"required,oneof=basic_care professional"`
	Professional *ProfessionalRequestData `json:"professional" binding:"required_if=AccountType professional"` // Magia do Gin
}
type ProfessionalRequestData struct {
	Kind               string  `json:"kind" binding:"required"`
	RegistrationNumber string  `json:"registration_number" binding:"required"`
	RegistrationIssuer string  `json:"registration_issuer" binding:"required"`
	RegistrationState  *string `json:"registration_state,omitempty"`
}

func (h *UserHandler) Register(c *gin.Context) {
	// 0) Wiring safety
	if h.svc == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "serviço indisponível",
		})
		return
	}

	// 1. Auth (Infra)
	identity, ok := middleware.GetIdentity(c)
	if !ok {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}
	// 2. Bind & Validate Formato (Infra)
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "payload inválido",
			Cause:   err,
		})
		return
	}
	// 3. Normalização de Email (Regra de Interface)
	email := req.Email
	if identity.Email != "" {
		email = identity.Email // Token tem prioridade
	}
	email = strings.TrimSpace(strings.ToLower(email))

	// 4) Dispatcher / role (Interface)
	accountType := user.AccountType(strings.ToLower(strings.TrimSpace(req.AccountType))).Normalize()
	if accountType == "" || !accountType.IsValid() || accountType == user.AccountTypeAdmin {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INVALID_ENUM_VALUE,
			Message: "account_type inválido",
		})
		return
	}

	// 5) Parse de campos que são “domínio” (você pode mover isso pro service depois)
	birthDate, err := common.ParseBirthDate(req.BirthDate)
	if err != nil {
		// O ParseBirthDate já retorna erro com %w (shared.ErrInvalidBirthDate),
		// então aqui basta traduzir para contrato.
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "data de nascimento inválida",
			Cause:   err,
		})
		return
	}

	input := userport.RegisterInput{
		Provider:    identity.Provider,
		Subject:     identity.Subject,
		Email:       email,
		FullName:    req.FullName,
		AccountType: accountType,
		BirthDate:   birthDate,
		CPF:         req.CPF,
		Phone:       req.Phone,
	}

	if accountType == user.AccountTypeProfessional {
		// Safe-check (mesmo que binder valide)
		if req.Professional == nil {
			httperrors.WriteError(c, &apperr.AppError{
				Code:    apperr.REQUIRED_FIELD_MISSING,
				Message: "dados profissionais são obrigatórios",
			})
			return
		}

		kind := professional.Kind(strings.ToLower(strings.TrimSpace(req.Professional.Kind))).Normalize()
		if !kind.IsValid() {
			httperrors.WriteError(c, &apperr.AppError{
				Code:    apperr.INVALID_ENUM_VALUE,
				Message: "professional.kind invÇ­lido",
			})
			return
		}

		input.Professional = &userport.ProfessionalRegistrationInput{
			Kind:               kind,
			RegistrationNumber: req.Professional.RegistrationNumber,
			RegistrationIssuer: req.Professional.RegistrationIssuer,
			RegistrationState:  req.Professional.RegistrationState,
		}
	}

	// 7) Chama serviço (Application)
	created, err := h.svc.Register(c.Request.Context(), input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticaÇõÇœo necessÇ­ria",
		})
		return
	}

	c.JSON(http.StatusOK, currentUser)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("user_update")

	// 0) Wiring safety
	if h.svc == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "serviço indisponível",
		})
		return
	}

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "payload inválido",
			Cause:   err,
		})
		return
	}

	input := userport.UpdateInput{
		UserID: currentUser.ID,
	}

	if req.FullName != nil {
		name := strings.TrimSpace(*req.FullName)
		input.FullName = &name
	}

	if req.BirthDate != nil {
		parsed, err := common.ParseBirthDate(*req.BirthDate)
		if err != nil {
			httperrors.WriteError(c, &apperr.AppError{
				Code:    apperr.VALIDATION_FAILED,
				Message: "data de nascimento inválida",
				Cause:   err,
			})
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

	updated, err := h.svc.Update(c.Request.Context(), input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	log.Info("user_updated")
	c.JSON(http.StatusOK, updated)
}
