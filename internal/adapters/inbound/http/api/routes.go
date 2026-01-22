package api

import (
	"sonnda-api/internal/adapters/inbound/http/api/handlers"
	"sonnda-api/internal/adapters/inbound/http/api/handlers/patient"
	"sonnda-api/internal/adapters/inbound/http/api/handlers/user"
	"sonnda-api/internal/adapters/inbound/http/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	registrationMiddleware *middleware.RegistrationMiddleware,
	userHandler *user.Handler,
	patientHandler *patient.PatientHandler,
	labsHandler *handlers.LabsHandler,

) {

	api := r.Group("/api/v1")

	// ---------------------------------------------------------------------
	// NÍVEL 1: Público (Health Check, Webhooks, Docs)
	// Ninguém precisa de token aqui.
	// ---------------------------------------------------------------------
	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API Documentation would be here."})
	})

	// ---------------------------------------------------------------------
	// NÍVEL 2: Autenticado (Tem Token do Firebase)
	// Aqui o cara provou que é dono do e-mail, mas talvez não tenha cadastro no banco.
	// ---------------------------------------------------------------------

	authenticated := api.Group("")
	authenticated.Use(authMiddleware.Authenticate())
	{
		// Rota de Cadastro (Onboarding)
		authenticated.POST("/api/v1/register",
			userHandler.Register,
		)
	}

	// ---------------------------------------------------------------------
	// NÍVEL 3: Registrado (Tem Token + Tem usuário no Banco)
	// Aqui é a área logada do app (Pacientes, Prontuários).
	// ---------------------------------------------------------------------

	protected := api.Group("")
	protected.Use(authMiddleware.Authenticate(), registrationMiddleware.RequireRegisteredUser())
	{
		//Perfil de usuário
		me := protected.Group("/me")
		{
			me.GET("", userHandler.GetUser)
			me.PUT("", userHandler.UpdateUser)
			me.DELETE("", userHandler.HardDeleteUser)
			me.GET("/patients", userHandler.ListMyPatients)
		}

		//Pacientes
		patients := protected.Group("/patients")
		{
			//Cria paciente
			patients.POST("", patientHandler.Create)
			patients.GET("", patientHandler.ListPatients)
			//Lista pacientes que o usuário tem acesso.
			//patients.GET("", patientHandler.ListAcessiblePatients)

			//Dados básicos do paciente
			patients.GET("/:id", patientHandler.GetPatient)

			labs := patients.Group("/:id/labs")
			{
				labs.GET("", labsHandler.ListLabs)
				labs.POST("/upload", labsHandler.UploadAndProcessLabs)

				labs.GET("/full", labsHandler.ListFullLabs)
			}

		}
	}
}
