package api

import (
	"net/http"

	"sonnda-api/internal/http/api/handlers/labs"
	"sonnda-api/internal/http/api/handlers/patient"
	"sonnda-api/internal/http/api/handlers/user"
	"sonnda-api/internal/http/middleware"

	"sonnda-api/assets"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	registrationMiddleware *middleware.RegistrationMiddleware,
	userHandler *user.UserHandler,
	patientHandler *patient.PatientHandler,
	labsHandler *labs.LabsHandler,
) {
	r.GET("/favicon.ico", func(c *gin.Context) {
		b, _ := assets.FS.ReadFile("favicon.ico")
		c.Data(http.StatusOK, "image/x-icon", b)
	})

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
		// Essa é a rota mágica!
		// O App chama ela logo após o login.
		// Se der 404/403, o App sabe que tem que mostrar a tela de cadastro.
		// Se der 200, o App vai pra Home.
		authenticated.GET("/check-registration",
			registrationMiddleware.RequireRegisteredUser(),
			userHandler.GetUser,
		)

		// Rota de Cadastro (Onboarding)
		// O middleware RequireUnregisteredUser garante que quem já tem cadastro
		// não consiga criar de novo (evita duplicidade).
		authenticated.POST("/register",
			registrationMiddleware.RequireUnregisteredUser(),
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
		}

		//Pacientes
		patients := protected.Group("/patients")
		{
			//Cria paciente
			patients.POST("", patientHandler.Create)
			//Lista pacientes que o usuário tem acesso.
			//patients.GET("", patientHandler.ListAcessiblePatients)

			//Dados básicos do paciente
			patients.GET("", patientHandler.GetByID)

			//Memberships do paciente
			//patients.GET("/members", patientHandler.ListPatientMembers)
			//patients.POST("/members", patientHandler.LinkToPatientByID)
			//patients.DELETE("/members", patientHandler.UnlinkFromPatientByID)
			//Atualiza paciente
			//patients.PUT("", patientHandler.UpdatePatientByID)
			//Deleta paciente
			//patients.DELETE("", patientHandler.DeletePatientByID)

			professionalRoutes := patients.Group("/medical-records")
			{
				professionalRoutes.GET("/labs", labsHandler.ListFullLabs)
				professionalRoutes.POST("/labs/upload", labsHandler.UploadAndProcessLabs)
				professionalRoutes.GET("/labs/summary", labsHandler.ListLabs)
			}

		}
	}
}
