package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/app/apperr"

	"sonnda-api/internal/app/interfaces/repositories"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/http/api/handlers/common"
	httperrors "sonnda-api/internal/http/errors"
	"sonnda-api/internal/http/middleware"
)

type UserHandler struct {
	svc usersvc.UserService
}

func NewUserHandler(
	svc usersvc.UserService,
	patientAccessRepo repositories.PatientAccessRepository,
) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

type RegisterRequest struct {
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
type UpdateUserRequest struct {
	FullName  *string `json:"full_name,omitempty"`
	BirthDate *string `json:"birth_date,omitempty" binding:"required,datetime=2006-01-02"`
	CPF       *string `json:"cpf,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}

func (h *UserHandler) Register(c *gin.Context) {

	// 1. Auth (Infra)
	identity, ok := middleware.GetIdentity(c)
	if !ok {
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}
	// 2. Bind & Validate Formato (Infra)
	var req RegisterRequest
	if err := httperrors.BindJSON(c, &req); err != nil {
		httperrors.WriteError(c, err)
		return
	}
	// 3. Parse de campos que são "domínio"
	// 3.1 AccountType já está validado pelo binding oneof
	accountType := user.AccountType(req.AccountType).Normalize()

	// BirthDate format já foi validado pelo binding datetime
	birthDate, err := common.ParseBirthDate(req.BirthDate)
	if err != nil {
		// Falha de conversão (muito rara, pois formato foi validado)
		httperrors.WriteError(c, apperr.Validation("data de nascimento inválida",
			apperr.Violation{
				Field:  "birth_date",
				Reason: "invalid_format",
			}))
		return
	}

	input := usersvc.UserRegisterInput{
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

		input.Professional = &usersvc.ProfessionalRegisterInput{
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
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	c.JSON(http.StatusOK, currentUser)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
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
	}

	if req.BirthDate != nil {
		// BirthDate format já foi validado pelo binding datetime
		// Apenas converter string → time.Time
		parsed, _ := common.ParseBirthDate(*req.BirthDate)
		input.BirthDate = &parsed
	}

	updated, err := h.svc.Update(c.Request.Context(), input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *UserHandler) HardDeleteUser(c *gin.Context) {
	// 0) Wiring safety
	if h.svc == nil {
		httperrors.WriteError(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, apperr.Unauthorized("autenticação necessária"))
		return
	}

	err := h.svc.Delete(c.Request.Context(), currentUser.ID)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
