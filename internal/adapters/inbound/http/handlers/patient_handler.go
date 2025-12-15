package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/usecases/patient"
	applog "sonnda-api/internal/logger"
)

type PatientHandler struct {
	createUC *patient.CreatePatientUseCase
	getUC    *patient.GetPatientUseCase
	updateUC *patient.UpdatePatientUseCase
	listUC   *patient.ListPatientsUseCase
}

func NewPatientHandler(
	createUC *patient.CreatePatientUseCase,
	getUC *patient.GetPatientUseCase,
	updateUC *patient.UpdatePatientUseCase,
	listUC *patient.ListPatientsUseCase,
) *PatientHandler {
	return &PatientHandler{
		createUC: createUC,
		getUC:    getUC,
		updateUC: updateUC,
		listUC:   listUC,
	}
}

func (h *PatientHandler) CreateByAuthenticatedPatient(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_create_by_authenticated_patient")

	var req patient.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	user, _ := middleware.RequireUser(c)
	cmd := patient.CreatePatientCommand{
		AppUserID:            &user.ID,
		CreatePatientRequest: req,
	}

	//impede crianção de paciente para outra pessoa
	if cmd.AppUserID != nil && *cmd.AppUserID != user.ID {
		RespondError(c, log, http.StatusForbidden, "forbidden", nil)
		return
	}

	output, err := h.createUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	log.Info("patient_created", "patient_id", output.ID)
	c.JSON(http.StatusCreated, output)

}

func (h *PatientHandler) CreateByProfessional(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_create_by_professional")

	var req patient.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	cmd := patient.CreatePatientCommand{
		AppUserID:            nil,
		CreatePatientRequest: req,
	}

	// Continua com criação
	output, err := h.createUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (h *PatientHandler) GetMyProfile(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_get_my_profile")

	// 1. Recupera o usuário autenticado
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		RespondError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// 2. Chama um método específico do use case
	p, err := h.getUC.ExecuteByUserID(c.Request.Context(), user.ID)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_patient_id", err)
		return

	}

	p, err := h.getUC.ExecuteByID(c.Request.Context(), id)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) GetByCPF(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	cpf := c.Param("cpf")
	if cpf == "" {
		RespondError(c, log, http.StatusBadRequest, "missing_cpf", nil)
		return
	}

	p, err := h.getUC.ExecuteByCPF(c.Request.Context(), cpf)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateByID(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_update_by_id")

	//Recuperar o usuário do contexto
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		RespondError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	var input patient.PatientChanges
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	p, err := h.updateUC.ExecuteByID(c.Request.Context(), user, id, input)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateByCPF(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_update_by_cpf")

	//Recuperar o usuário do contexto
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		RespondError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	cpf := c.Param("cpf")
	if cpf == "" {
		RespondError(c, log, http.StatusBadRequest, "missing_cpf", nil)
		return
	}

	var input patient.PatientChanges
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	p, err := h.updateUC.ExecuteByCPF(c.Request.Context(), user, cpf, input)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) List(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_list")
	list, err := h.listUC.Execute(c.Request.Context(), 100, 0)
	if err != nil {
		RespondDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

/* ============================================================
   Error helpers (centraliza log + resposta)
   ============================================================ */
