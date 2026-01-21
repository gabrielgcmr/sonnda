package web

import (
	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/web/handlers"
)

func SetupRoutes(r *gin.Engine) {
	h := handlers.NewHomeHandler()

	r.GET("/", h.Home)
	r.GET("/partials/counter", h.CounterPartial)
}
