package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/domain"
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

	var input patient.CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		handlePatientError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	//impede crianção de paciente para outra pessoa
	if input.AppUserID != nil && *input.AppUserID != user.ID {
		handlePatientError(c, log, http.StatusForbidden, "forbidden", nil)
		return
	}

	input.AppUserID = &user.ID

	output, err := h.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, log, err)
		return
	}

	log.Info("patient_created", "patient_id", output.ID)
	c.JSON(http.StatusCreated, output)

}

func (h *PatientHandler) CreateByProfessional(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_create_by_professional")

	var input patient.CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		handlePatientError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Protege contra tentativas de injetar app_user_id
	if input.AppUserID != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_input", nil)
		return
	}

	// Continua com criação
	output, err := h.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, log, err)
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
		handlePatientError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// 2. Chama um método específico do use case
	p, err := h.getUC.ExecuteByUserID(c.Request.Context(), user.ID)
	if err != nil {
		handleServiceError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_patient_id", err)
		return

	}

	p, err := h.getUC.ExecuteByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) GetByCPF(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())

	cpf := c.Param("cpf")
	if cpf == "" {
		handlePatientError(c, log, http.StatusBadRequest, "missing_cpf", nil)
		return
	}

	p, err := h.getUC.ExecuteByCPF(c.Request.Context(), cpf)
	if err != nil {
		handleServiceError(c, log, err)
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
		handlePatientError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	var input patient.PatientChanges
	if err := c.ShouldBindJSON(&input); err != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	p, err := h.updateUC.ExecuteByID(c.Request.Context(), user, id, input)
	if err != nil {
		handleServiceError(c, log, err)
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
		handlePatientError(c, log, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	cpf := c.Param("cpf")
	if cpf == "" {
		handlePatientError(c, log, http.StatusBadRequest, "missing_cpf", nil)
		return
	}

	var input patient.PatientChanges
	if err := c.ShouldBindJSON(&input); err != nil {
		handlePatientError(c, log, http.StatusBadRequest, "invalid_input", err)
		return
	}

	p, err := h.updateUC.ExecuteByCPF(c.Request.Context(), user, cpf, input)
	if err != nil {
		handleServiceError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) List(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("patient_list")
	list, err := h.listUC.Execute(c.Request.Context(), 100, 0)
	if err != nil {
		handleServiceError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

/* ============================================================
   Error helpers (centraliza log + resposta)
   ============================================================ */

func handleServiceError(c *gin.Context, log *slog.Logger, err error) {
	// Mapeia domínio -> HTTP + loga com o nível adequado
	switch err {
	case domain.ErrCPFAlreadyExists:
		log.Warn("service_error", "error", "cpf_already_exists")
		c.JSON(http.StatusConflict, gin.H{"error": "cpf_already_exists"})
	case domain.ErrPatientNotFound:
		// “not found” é esperado às vezes → Warn ok (ou Info, se preferir)
		log.Warn("service_error", "error", "patient_not_found")
		c.JSON(http.StatusNotFound, gin.H{"error": "patient_not_found"})
	case domain.ErrInvalidBirthDate:
		log.Warn("service_error", "error", "invalid_birth_date")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_birth_date"})
	case domain.ErrPatientTooYoung:
		log.Warn("service_error", "error", "patient_too_young")
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_too_young"})
	default:
		log.Error("service_error", "error", "server_error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"details": err.Error(),
		})
	}
}

func handlePatientError(c *gin.Context, log *slog.Logger, status int, code string, err error) {
	// refinamento por status:
	// 401/403 -> Info
	// 4xx -> Warn
	// 5xx -> Error
	switch {
	case status >= 500:
		if err != nil {
			log.Error("handler_error", "status", status, "error", code, "err", err)
		} else {
			log.Error("handler_error", "status", status, "error", code)
		}
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		if err != nil {
			log.Info("handler_error", "status", status, "error", code, "err", err)
		} else {
			log.Info("handler_error", "status", status, "error", code)
		}
	default:
		if err != nil {
			log.Warn("handler_error", "status", status, "error", code, "err", err)
		} else {
			log.Warn("handler_error", "status", status, "error", code)
		}
	}

	// resposta
	if err != nil {
		c.JSON(status, gin.H{
			"error":   code,
			"details": err.Error(),
		})
		return
	}
	c.JSON(status, gin.H{"error": code})
}
