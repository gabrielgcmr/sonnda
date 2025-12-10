package http

import (
	"sonnda-api/internal/adapters/inbound/http/handlers"
	"sonnda-api/internal/adapters/inbound/http/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	userHandler *handlers.UserHandler,
	patientHandler *handlers.PatientHandler,
	labsHandler *handlers.LabsHandler,

) *gin.Engine {

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	api := r.Group("/api/v1")

	public := api.Group("/public")
	{
		public.GET("/docs", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "API Documentation would be here."})
		})
	}

	api.POST("/register", userHandler.CreateUserFromIdentity)

	protected := api.Group("")
	protected.Use(authMiddleware.Authenticate())
	me := protected.Group("/me")
	{
		me.GET("", userHandler.GetCurrentUser)
		me.GET("/profile",
			middleware.RequirePatient(),
			patientHandler.GetMyProfile,
		)
		me.POST("/patient",
			middleware.RequirePatient(),
			patientHandler.Create,
		)
		me.GET("labs",
			middleware.RequirePatient(),
			labsHandler.ListMyLabReports,
		)
		me.POST("labs/upload",
			middleware.RequirePatient(),
			labsHandler.UploadAndProcessLabReport,
		)
		//me.GET("/doctor", userHandler.GetCurrentDoctor)
	}
	// Rotas de pacientes
	patients := protected.Group("/patients")
	{
		// Qualquer usuário autenticado pode listar
		patients.GET("", patientHandler.List)

		// Apenas doctors podem criar pacientes
		patients.POST("",
			middleware.RequireDoctor(),
			patientHandler.Create,
		)

		// Buscar por ID (doctors only)
		patients.GET("/:id",
			middleware.RequireDoctor(),
			patientHandler.GetByID,
		)

		// Atualizar (doctors only)
		patients.PUT("/:id",
			middleware.RequireDoctor(),
			patientHandler.UpdateByID,
		)

		// Upload de laudo (doctors only)
		patients.POST("/:id/labs/upload",
			middleware.RequireDoctor(),
			labsHandler.UploadAndProcessLabReport,
		)

		// Rotas específicas de doctors

	}

	doctors := protected.Group("/doctors")
	doctors.Use(middleware.RequireDoctor())
	{
		// Exemplo: rotas específicas de médicos
	}

	return r
}
