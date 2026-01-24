// internal/adapters/inbound/http/web/handlers/home_handler.go
package handlers

import (
	"net/http"
	"sonnda-api/internal/adapters/inbound/http/shared/httpctx"
	"sonnda-api/internal/adapters/inbound/http/web/assets/templates/pages"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	counter atomic.Int64
}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) Home(c *gin.Context) {
	renderHomePage(c)
}

func renderHomePage(c *gin.Context) {
	// 1) Procura identity no contexto
	currentUser := httpctx.MustGetCurrentUser(c)

	//2)
	vm := pages.HomeViewModel{
		UserName: currentUser.FullName,
		Role:     string(currentUser.AccountType),
		Patients: nil, //TODO: buscar pacientes
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")

	// Render do templ no writer do Gin
	if err := pages.Home(vm).Render(c.Request.Context(), c.Writer); err != nil {
		// Aqui dá pra usar seu httperr/apperr se quiser, mas pra web geralmente:
		c.Status(http.StatusInternalServerError)
	}
}
