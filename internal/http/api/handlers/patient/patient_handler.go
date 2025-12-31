package patient

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	applog "sonnda-api/internal/app/observability"
	patientsvc "sonnda-api/internal/app/services/patient"

	"sonnda-api/internal/http/api/handlers/common"
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

	user, ok := middleware.GetCurrentUser(c)
	if !ok || user == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req CreatePatientRequest
	// 1. Bind do request
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_input", err)
		return
	}

	// 3. Parsing / normalização de fronteira
	birthDate, err := ParseBirthDate(req.BirthDate)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	gender, err := ParseGender(req.Gender)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	race, err := ParseRace(req.Race)
	if err != nil {
		common.RespondAppError(c, err)
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
		common.RespondAppError(c, err)
		return
	}

	c.Header("Location", "/patients/"+p.ID.String())
	c.JSON(http.StatusCreated, gin.H{
		"id": p.ID.String(),
	})
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	id := c.Param("id")
	if id == "" {
		common.RespondError(c, http.StatusBadRequest, "missing_patient_id", nil)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	p, err := h.svc.GetByID(c.Request.Context(), currentUser, parsedID)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateByID(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_update_by_id")

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	id := c.Param("id")
	if id == "" {
		common.RespondError(c, http.StatusBadRequest, "missing_patient_id", nil)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	var input patientsvc.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_input", err)
		return
	}

	p, err := h.svc.UpdateByID(c.Request.Context(), currentUser, parsedID, input)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) List(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_list")

	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	list, err := h.svc.List(c.Request.Context(), currentUser, 100, 0)
	if err != nil {
		common.RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

/* ============================================================
   Error helpers (centraliza log + resposta)
   ============================================================ */
