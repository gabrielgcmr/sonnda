package handlers

import (
	"net/http"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/usecases/patient"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var input patient.CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	//impede crianção de paciente para outra pessoa
	if input.AppUserID != nil && *input.AppUserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "você não tem permissão para criar paciente para outro paciente",
		})
		return
	}

	input.AppUserID = &user.ID

	output, err := h.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)

}

func (h *PatientHandler) CreateByProfessional(c *gin.Context) {
	var input patient.CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	// Protege contra tentativas de injetar app_user_id
	if input.AppUserID != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_input",
			"message": "médicos não devem informar app_user_id",
		})
		return
	}

	// Continua com criação
	output, err := h.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (h *PatientHandler) GetMyProfile(c *gin.Context) {
	// 1. Recupera o usuário autenticado
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	// 2. Chama um método específico do use case
	p, err := h.getUC.ExecuteByUserID(c.Request.Context(), user.ID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
		return

	}

	p, err := h.getUC.ExecuteByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) GetByCPF(c *gin.Context) {
	cpf := c.Param("cpf")

	p, err := h.getUC.ExecuteByCPF(c.Request.Context(), cpf)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateByID(ctx *gin.Context) {
	//Recuperar o usuário do contexto
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	idStr := ctx.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
		return
	}

	var input patient.PatientChanges
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	p, err := h.updateUC.ExecuteByID(ctx.Request.Context(), user, id, input)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, p)
}

func (h *PatientHandler) UpdateByCPF(ctx *gin.Context) {
	//Recuperar o usuário do contexto
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuário não autenticado",
		})
		return
	}

	cpf := ctx.Param("cpf")

	var input patient.PatientChanges
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	p, err := h.updateUC.ExecuteByCPF(ctx.Request.Context(), user, cpf, input)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, p)
}

func (h *PatientHandler) List(ctx *gin.Context) {
	list, err := h.listUC.Execute(ctx.Request.Context(), 100, 0)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, list)
}

func handleServiceError(c *gin.Context, err error) {
	switch err {
	case domain.ErrCPFAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "cpf_already_exists"})
	case domain.ErrPatientNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "patient_not_found"})
	case domain.ErrInvalidBirthDate:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_birth_date"})
	case domain.ErrPatientTooYoung:
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_too_young"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"details": err.Error(),
		})
	}
}
