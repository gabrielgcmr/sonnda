// internal/api/routes.go
package api

import (
	"log/slog"
	"net/http"

	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	"github.com/gabrielgcmr/sonnda/internal/api/handlers/patient"
	"github.com/gabrielgcmr/sonnda/internal/api/handlers/user"
	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gabrielgcmr/sonnda/internal/config"

	"github.com/gin-gonic/gin"
)

type Infra struct {
	Logger *slog.Logger
	Config *config.Config
}

type Deps struct {
	API *APIDependencies
}

type APIDependencies struct {
	AuthMiddleware         *middleware.AuthMiddleware
	RegistrationMiddleware *middleware.RegistrationMiddleware
	UserHandler            *user.Handler
	PatientHandler         *patient.PatientHandler
	LabsHandler            *handlers.LabsHandler
}

const rootAPIName = "sonnda"

// rootAPIVersion is overridden via -ldflags in build/release pipelines.
var rootAPIVersion = "1.0.0"

type RootResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Docs        string `json:"docs"`
	OpenAPI     string `json:"openapi"`
	Health      string `json:"health"`
	Ready       string `json:"ready"`
}

func NewRouter(infra Infra, deps Deps) *gin.Engine {
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

	// ---- Rotas ----
	registerRootRoute(r, infra.Config)
	SetupRoutes(r, deps.API)

	return r
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
		// Rota de Cadastro (Onboarding)
		auth.POST("/register", deps.UserHandler.Register)
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

func registerRootRoute(r gin.IRouter, cfg *config.Config) {
	environment := "dev"
	if cfg != nil && cfg.Env != "" {
		environment = cfg.Env
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, RootResponse{
			Name:        rootAPIName,
			Version:     rootAPIVersion,
			Environment: environment,
			Docs:        "/docs",
			OpenAPI:     "/openapi.yaml",
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
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapiSpec)
	})
}
