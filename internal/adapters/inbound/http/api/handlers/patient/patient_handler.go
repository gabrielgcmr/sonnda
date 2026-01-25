package patient

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sonnda-api/internal/app/apperr"
	applog "sonnda-api/internal/app/observability"
	patientsvc "sonnda-api/internal/app/services/patient"

	"sonnda-api/internal/adapters/inbound/http/api/apierr"
	"sonnda-api/internal/adapters/inbound/http/api/binder"
	"sonnda-api/internal/adapters/inbound/http/api/handlers"
	"sonnda-api/internal/adapters/inbound/http/shared/httpctx"
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

	user, ok := httpctx.GetCurrentUser(c)
	if !ok || user == nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		return
	}

	var req CreatePatientRequest
	// 1. Bind do request
	if err := binder.BindJSON(c, &req); err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	// 3. Parsing / normalização de fronteira
	birthDate, err := handlers.ParseBirthDate(req.BirthDate)
	if err != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "data de nascimento inválida",
			Cause:   err,
		})
		return
	}

	gender, err := ParseGender(req.Gender)
	if err != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "gênero inválido",
			Cause:   err,
		})
		return
	}

	race, err := ParseRace(req.Race)
	if err != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
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
		apierr.ErrorResponder(c, err)
		return
	}

	c.Header("Location", "/patients/"+p.ID.String())
	c.JSON(http.StatusCreated, gin.H{
		"id": p.ID.String(),
	})
}

func (h *PatientHandler) GetPatient(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)

	id := c.Param("id")
	if id == "" {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return
	}

	p, err := h.svc.Get(c.Request.Context(), currentUser, parsedID)
	if err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)

	id := c.Param("id")
	if id == "" {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return
	}

	var input patientsvc.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "payload inválido",
			Cause:   err,
		})
		return
	}

	p, err := h.svc.Update(c.Request.Context(), currentUser, parsedID, input)
	if err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) ListPatients(c *gin.Context) {
	if h == nil || h.svc == nil {
		apierr.ErrorResponder(c, apperr.Internal("serviço indisponível", nil))
		return
	}

	currentUser := httpctx.MustGetCurrentUser(c)

	list, err := h.svc.ListMyPatients(c.Request.Context(), currentUser, 100, 0)
	if err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *PatientHandler) HardDeletePatient(c *gin.Context) {
	currentUser := httpctx.MustGetCurrentUser(c)

	id := c.Param("id")
	if id == "" {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "patient_id é obrigatório",
		})
		return
	}

	parsedID, parseErr := uuid.Parse(id)
	if parseErr != nil {
		apierr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.INVALID_FIELD_FORMAT,
			Message: "patient_id inválido",
			Cause:   parseErr,
		})
		return
	}

	if err := h.svc.HardDelete(c.Request.Context(), currentUser, parsedID); err != nil {
		apierr.ErrorResponder(c, err)
		return
	}

	c.Status(http.StatusNoContent)

}
