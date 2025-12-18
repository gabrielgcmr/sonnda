package http

import (
	"embed"
	"net/http"
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
	var faviconFS embed.FS

	r.GET("/favicon.ico", func(c *gin.Context) {
		b, _ := faviconFS.ReadFile("favicon.ico")
		c.Data(http.StatusOK, "image/x-icon", b)
	})

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
			patientHandler.CreateByAuthenticatedPatient,
		)

		//me.GET("/doctor", userHandler.GetCurrentDoctor)
	}
	//Upload Labs
	protected.POST("/:patientID/labs/upload",
		middleware.RequirePatient(),
		labsHandler.UploadAndProcessLabs,
	)
	protected.GET("/:patientID/labs",
		middleware.RequirePatient(),
		labsHandler.ListFullLabs,
	)
	protected.GET("/:patientID/labs/summary",
		middleware.RequirePatient(),
		labsHandler.ListLabs,
	)
	//TODO: GET /patients/{patientID}/labs?from=2025-01-01&to=2025-01-31

	// Rotas para profissionais de saúde
	patients := protected.Group("/patients")
	{
		// Qualquer usuário autenticado pode listar
		patients.GET("", patientHandler.List)

		// Apenas doctors podem criar pacientes
		patients.POST("",
			middleware.RequireProfessional(),
			patientHandler.CreateByProfessional,
		)

		// Buscar por ID (doctors only)
		patients.GET("/:id",
			middleware.RequireProfessional(),
			patientHandler.GetByID,
		)

		// Atualizar (doctors only)
		patients.PUT("/:id",
			middleware.RequireProfessional(),
			patientHandler.UpdateByID,
		)

		// Upload de laudo (doctors only)
		patients.POST("/:id/labs/upload",
			middleware.RequireProfessional(),
			labsHandler.UploadAndProcessLabs,
		)

		// Rotas específicas de doctors

	}

	doctors := protected.Group("/doctors")
	doctors.Use(middleware.RequireProfessional())
	{
		// Exemplo: rotas específicas de médicos
	}

	return r
}
