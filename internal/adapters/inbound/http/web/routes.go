// internal/adapters/inbound/http/web/routes.go
package web

import (
	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/web/handlers"
	webmw "sonnda-api/internal/adapters/inbound/http/web/middleware"
)

type WebDependencies struct {
	HomeHandler     *handlers.HomeHandler
	AuthHandler     *handlers.AuthHandler
	SessionHandler  *handlers.SessionHandler
	WebAuth         *webmw.AuthMiddleware
	WebRegistration *webmw.RegistrationMiddleware
}

func SetupRoutes(
	r *gin.Engine,
	deps WebDependencies,
) {
	// ---------------------------------------------------------------------
	// NÍVEL 1: Público
	// ---------------------------------------------------------------------

	r.GET("/login", deps.AuthHandler.Login)
	r.GET("/register", deps.AuthHandler.Register)

	// Endpoints de sessão (fluxo de auth no browser)
	r.POST("/auth/session", deps.SessionHandler.CreateSession)
	r.DELETE("/auth/session", deps.SessionHandler.DeleteSession)
	r.POST("/auth/logout", deps.SessionHandler.Logout)
	r.GET("/auth/session", deps.SessionHandler.GetSession)
	r.POST("/auth/session/refresh", deps.SessionHandler.RefreshSession)

	// ---------------------------------------------------------------------
	// NÍVEL 2: Autenticado (cookie válido)
	// ---------------------------------------------------------------------

	authenticated := r.Group("")
	authenticated.Use(deps.WebAuth.RequireSession())
	{
		// Exemplo: páginas que só precisam estar logado,
		// mas não necessariamente ter cadastro local (se isso existir no seu fluxo).
		// authenticated.GET("/welcome", homeHandler.Welcome)
	}

	// ---------------------------------------------------------------------
	// NÍVEL 3: Registrado (cookie + user local)
	// ---------------------------------------------------------------------

	registered := r.Group("")
	registered.Use(
		deps.WebAuth.RequireSession(),
		deps.WebRegistration.RequireRegisteredUser(),
	)
	{
		registered.GET("/", deps.HomeHandler.Home)

		// E depois: /patients (HTML), etc.
		// registered.GET("/patients", patientsPageHandler.List)
	}
}
