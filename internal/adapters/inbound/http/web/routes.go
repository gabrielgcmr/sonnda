// internal/adapters/inbound/http/web/routes.go
package web

import (
	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/web/handlers"
	"sonnda-api/internal/app/config"
	"sonnda-api/internal/domain/ports/integration"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, identityService integration.IdentityService) {
	h := handlers.NewHomeHandler()
	authHandler := handlers.NewAuthHandler(cfg)
	sessionHandler := handlers.NewSessionHandler(identityService)

	// Home
	r.GET("/", h.Home)
	r.GET("/partials/counter", h.CounterPartial)

	// Authentication pages
	r.GET("/login", authHandler.Login)
	r.GET("/register", authHandler.Register)

	// Session management (API endpoints for auth flow)
	r.POST("/auth/session", sessionHandler.CreateSession)
	r.DELETE("/auth/session", sessionHandler.DeleteSession)
	r.POST("/auth/logout", sessionHandler.Logout)
	r.GET("/auth/session", sessionHandler.GetSession)
	r.POST("/auth/session/refresh", sessionHandler.RefreshSession)
}
