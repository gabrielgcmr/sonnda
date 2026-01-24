// internal/adapters/inbound/http/web/handlers/home_handler.go
package handlers

import (
	"net/http"
	"sonnda-api/internal/adapters/inbound/http/shared/httpctx"
	"sonnda-api/internal/adapters/inbound/http/web/templates/components"
	"sonnda-api/internal/adapters/inbound/http/web/templates/pages"
	patientsvc "sonnda-api/internal/app/services/patient"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	patientService patientsvc.Service
	counter        atomic.Int64
}

func NewHomeHandler(patientService patientsvc.Service) *HomeHandler {
	return &HomeHandler{
		patientService: patientService,
	}
}

func (h *HomeHandler) Home(c *gin.Context) {
	h.renderHomePage(c)
}

func (h *HomeHandler) renderHomePage(c *gin.Context) {
	// 1) Procura identity no contexto
	currentUser := httpctx.MustGetCurrentUser(c)

	// 2) Busca pacientes
	// TODO: adicionar paginação
	patients, err := h.patientService.ListMyPatients(c.Request.Context(), currentUser, 10, 0)
	if err != nil {
		// Log erro, mas renderiza página sem pacientes por enquanto
		// ou redireciona para erro.
		// Vamos assumir lista vazia por segurança
		patients = nil
	}

	// 3) Converte para ViewModel
	var patientItems []components.PatientItem
	for _, p := range patients {
		patientItems = append(patientItems, components.PatientItem{
			ID:   p.ID.String(),
			Name: p.Name,
		})
	}

	vm := pages.HomeViewModel{
		UserName: currentUser.FullName,
		Role:     string(currentUser.AccountType),
		Patients: patientItems,
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")

	// Render do templ no writer do Gin
	if err := pages.Home(vm).Render(c.Request.Context(), c.Writer); err != nil {
		// Aqui dá pra usar seu httperr/apperr se quiser, mas pra web geralmente:
		c.Status(http.StatusInternalServerError)
	}
}
