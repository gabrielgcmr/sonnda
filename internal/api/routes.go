// internal/api/routes.go
package api

import (
	"net/http"

	"github.com/gabrielgcmr/sonnda/internal/api/handlers"

	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	openapispec "github.com/gabrielgcmr/sonnda/internal/api/openapi"
	openapigen "github.com/gabrielgcmr/sonnda/internal/api/openapi/generated"

	"github.com/gin-gonic/gin"
)

type APIDependencies struct {
	AuthMiddleware         *middleware.AuthMiddleware
	RegistrationMiddleware *middleware.RegistrationMiddleware
	UserHandler            *handlers.UserHandler
	PatientHandler         *handlers.PatientHandler
	LabsHandler            *handlers.LabsHandler
}

type RootInfo struct {
	Name    string
	Version string
	Env     string
}

func SetupRoutes(
	r gin.IRouter,
	deps *APIDependencies,

) {
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/x-icon", faviconData)
	})
	// Rotas públicas (sem versão)
	registerPublicRoutes(r)
	registerDocsRoutes(r)
	registerOpenAPIRoute(r)

	v1 := r.Group("/v1")

	// ---------------------------------------------------------------------
	// NÍVEL 2: Autenticado (Tem Token do Supabase)
	// Aqui o cara provou que é dono do e-mail, mas talvez não tenha cadastro no banco.
	// ---------------------------------------------------------------------

	auth := v1.Group("")
	auth.Use(deps.AuthMiddleware.RequireBearer())
	{
		// Criação de usuário (Onboarding)
		// OpenAPI: POST /v1/me
		auth.POST("/me", deps.UserHandler.CreateUser)
		// Legacy: keep /v1/users for backwards-compat
		auth.POST("/users", deps.UserHandler.CreateUser)
	}

	// ---------------------------------------------------------------------
	// NÍVEL 3: Registrado (Tem Token + Tem usuário no Banco)
	// Aqui é a área logada do app (Pacientes, Prontuários).
	// ---------------------------------------------------------------------

	registered := v1.Group("")
	registered.Use(
		deps.AuthMiddleware.RequireBearer(),
		deps.RegistrationMiddleware.RequireRegisteredUser())
	{
		//Perfil de usuário
		me := registered.Group("/me")
		{
			me.GET("", deps.UserHandler.GetUser)
			me.PUT("", deps.UserHandler.UpdateUser)
			me.DELETE("", deps.UserHandler.HardDeleteUser)
			me.GET("/patients", deps.UserHandler.ListMyPatients)
		}

		//Pacientes
		patients := registered.Group("/patients")
		{
			//Cria paciente
			patients.POST("", deps.PatientHandler.Create)
			patients.GET("", deps.PatientHandler.ListPatients)
			//Lista pacientes que o usuário tem acesso.
			//patients.GET("", deps.PatientHandler.ListAcessiblePatients)

			//Dados básicos do paciente
			patients.GET("/:id", deps.PatientHandler.GetPatient)

			labs := patients.Group("/:id/labs")
			{
				labs.GET("", deps.LabsHandler.ListLabs)
				labs.POST("", deps.LabsHandler.UploadAndProcessLabs)
			}

		}
	}
}

func registerRootRoute(r gin.IRouter, info RootInfo) {
	environment := info.Env
	if environment == "" {
		environment = "dev"
	}
	name := info.Name
	if name == "" {
		name = "Sonnda API"
	}
	version := info.Version
	if version == "" {
		version = "dev"
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, openapigen.RootResponse{
			Name:        name,
			Version:     version,
			Environment: environment,
			Docs:        "/docs",
			Openapi:     "/openapi.yaml",
			Health:      "/healthz",
			Ready:       "/readyz",
		})
	})
}

func registerPublicRoutes(r gin.IRouter) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
}

func registerDocsRoutes(r gin.IRouter) {
	r.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", docsHTML)
	})
}

func registerOpenAPIRoute(r gin.IRouter) {
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapispec.OpenAPISpec)
	})
}
