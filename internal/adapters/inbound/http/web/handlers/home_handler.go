// internal/adapters/inbound/http/web/handlers/home_handler.go
package handlers

import (
	"net/http"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/adapters/inbound/http/web/pages"

	"github.com/gin-gonic/gin"
)

func HomeHandler(c *gin.Context) {
	// 1) Procura identity no contexto
	id, ok := middleware.GetIdentity(c)
	if !ok || id == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	//2)
	vm := pages.HomeViewModel{
		UserName: id.FullName,
		Role:     id.Role.String(),
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
