package patient

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sonnda-api/internal/app/apperr"
	applog "sonnda-api/internal/app/observability"
	patientsvc "sonnda-api/internal/app/services/patient"

	"sonnda-api/internal/http/api/handlers/common"
	httperrors "sonnda-api/internal/http/errors"
	"sonnda-api/internal/http/middleware"
)

type PatientHandler struct {
	svc patientsvc.Service
}

func NewPatientHandler(svc patientsvc.Service) *PatientHandler {
	return &PatientHandler{svc: svc}
}

func (h *PatientHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	log := applog.FromContext(ctx)
	log.Info("patient_create")

	if h == nil || h.svc == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "serviço indisponível",
		})
		return
	}

	user, ok := middleware.GetCurrentUser(c)
	if !ok || user == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}

	var req CreatePatientRequest
	// 1. Bind do request
	if err := httperrors.BindJSON(c, &req); err != nil {
		httperrors.WriteError(c, err)
		return
	}

	// 3. Parsing / normalização de fronteira
	birthDate, err := common.ParseBirthDate(req.BirthDate)
	if err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "data de nascimento inválida",
			Cause:   err,
		})
		return
	}

	gender, err := ParseGender(req.Gender)
	if err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INVALID_ENUM_VALUE,
			Message: "gênero inválido",
			Cause:   err,
		})
		return
	}

	race, err := ParseRace(req.Race)
	if err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INVALID_ENUM_VALUE,
			Message: "raça inválida",
			Cause:   err,
		})
		return
	}

	// 4. Montagem do input da aplicação
	input := patientsvc.CreateInput{
		CPF:       req.CPF,
		FullName:  req.FullName,
		BirthDate: birthDate,
		Gender:    gender,
		Race:      race,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
	}

	// 5. Execução do use case
	p, err := h.svc.Create(ctx, user, input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.Header("Location", "/patients/"+p.ID.String())
	c.JSON(http.StatusCreated, gin.H{
		"id": p.ID.String(),
	})
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	if h == nil || h.svc == nil {
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

	id := c.Param("id")
	if id == "" {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.REQUIRED_FIELD_MISSING,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INVALID_FIELD_FORMAT,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return
	}

	p, err := h.svc.GetByID(c.Request.Context(), currentUser, parsedID)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateByID(c *gin.Context) {
	if h == nil || h.svc == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "serviço indisponível",
		})
		return
	}

	log := applog.FromContext(c.Request.Context())
	log.Info("patient_update_by_id")

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.REQUIRED_FIELD_MISSING,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INVALID_FIELD_FORMAT,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return
	}

	var input patientsvc.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "payload inválido",
			Cause:   err,
		})
		return
	}

	p, err := h.svc.UpdateByID(c.Request.Context(), currentUser, parsedID, input)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) List(c *gin.Context) {
	if h == nil || h.svc == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "serviço indisponível",
		})
		return
	}

	log := applog.FromContext(c.Request.Context())
	log.Info("patient_list")

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}

	list, err := h.svc.List(c.Request.Context(), currentUser, 100, 0)
	if err != nil {
		httperrors.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

/* ============================================================
   Error helpers (centraliza log + resposta)
   ============================================================ */
