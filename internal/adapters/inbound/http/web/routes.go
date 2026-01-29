// internal/adapters/inbound/http/web/routes.go
package web

import (
	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/handlers"
	webmw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/middleware"
)

type WebDependencies struct {
	HomeHandler     *handlers.HomeHandler
	SessionHandler  *handlers.SessionHandler
	WebAuth         *webmw.AuthMiddleware
	WebRegistration *webmw.RegistrationMiddleware
}

func SetupRoutes(
	r gin.IRouter,
	deps WebDependencies,
) {
	// ---------------------------------------------------------------------
	// NÍVEL 1: Público
	// ---------------------------------------------------------------------

	r.GET("/login", deps.SessionHandler.Login)
	r.GET("/callback", deps.SessionHandler.Callback)
	r.GET("/logout", deps.SessionHandler.Logout)
	r.GET("/session", deps.SessionHandler.GetSession) //Debug

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
