// File: internal/adapters/inbound/http/router.go
package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/api"
	"sonnda-api/internal/adapters/inbound/http/api/handlers"
	"sonnda-api/internal/adapters/inbound/http/api/handlers/patient"
	"sonnda-api/internal/adapters/inbound/http/api/handlers/user"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/adapters/inbound/http/web"
	"sonnda-api/internal/domain/ports/integration"
)

type Infra struct {
	Logger          *slog.Logger
	IdentityService integration.IdentityService
}

type Dependencies struct {
	AuthMiddleware         *middleware.AuthMiddleware
	RegistrationMiddleware *middleware.RegistrationMiddleware
	UserHandler            *user.Handler
	PatientHandler         *patient.PatientHandler
	LabsHandler            *handlers.LabsHandler
}

func NewRouter(infra Infra, deps Dependencies) *gin.Engine {
	r := gin.New()

	logger := infra.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Middlewares globais (infra)
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(logger),
		middleware.Recovery(logger),
	)

	// Static assets (css, js, imagens)
	// Use relative path that works from project root (where air runs from)
	r.Static("/assets", "assets")

	// Templates HTML (HTMX)
	r.SetHTMLTemplate(mustLoadTemplates())

	// ---- Rotas ----
	web.SetupRoutes(r, infra.IdentityService)
	api.SetupRoutes(
		r,
		deps.AuthMiddleware,
		deps.RegistrationMiddleware,
		deps.UserHandler,
		deps.PatientHandler,
		deps.LabsHandler,
	)

	return r
}
