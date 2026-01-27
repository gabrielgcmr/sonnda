// internal/adapters/inbound/http/api/routes.go
package api

import (
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers/patient"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers/user"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/middleware"

	"github.com/gin-gonic/gin"
)

type APIDependencies struct {
	AuthMiddleware         *middleware.AuthMiddleware
	RegistrationMiddleware *middleware.RegistrationMiddleware
	UserHandler            *user.Handler
	PatientHandler         *patient.PatientHandler
	LabsHandler            *handlers.LabsHandler
}

func SetupRoutes(
	r *gin.Engine,
	deps *APIDependencies,

) {
	v1 := r.Group("/v1")

	// ---------------------------------------------------------------------
	// NÍVEL 1: Público (Health Check, Webhooks, Docs)
	// Ninguém precisa de token aqui.
	// ---------------------------------------------------------------------
	v1.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	v1.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API Documentation would be here."})
	})

	// ---------------------------------------------------------------------
	// NÍVEL 2: Autenticado (Tem Token do Firebase)
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
				labs.POST("/upload", deps.LabsHandler.UploadAndProcessLabs)
				labs.GET("/full", deps.LabsHandler.ListFullLabs)
			}

		}
	}
}
