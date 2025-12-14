package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"sonnda-api/internal/adapters/inbound/http"
	"sonnda-api/internal/adapters/inbound/http/handlers"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/adapters/outbound/auth"
	"sonnda-api/internal/adapters/outbound/authorization"
	"sonnda-api/internal/adapters/outbound/database/supabase"
	"sonnda-api/internal/adapters/outbound/external/documentai"
	"sonnda-api/internal/adapters/outbound/storage"
	"sonnda-api/internal/config"
	"sonnda-api/internal/core/usecases/labs"
	"sonnda-api/internal/core/usecases/patient"
	"sonnda-api/internal/core/usecases/user"
	"sonnda-api/internal/logger"
)

func main() {
	// 1. Carrega o contexto
	ctx := context.Background()
	// 2. Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}
	// 3. Carrega configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar configuração: ", err)
	}

	//4. Carrega logger
	appLogger := logger.New(logger.Config{
		Env:       cfg.Env,
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		AppName:   "sonnda-api",
		AddSource: cfg.Env == "dev",
	})
	slog.SetDefault(appLogger)

	// 4. Conectar db (Supabase via pgxpool)
	dbClient, err := supabase.NewClient(config.SupabaseConfig(*cfg))
	if err != nil {
		log.Fatalf("falha ao criar client do supabase: %v", err)
	}
	defer dbClient.Close()

	//5. Conectando outros serviços
	//5.1 Storage Service (GCS)
	storageService, err := storage.NewStorageAdapter(ctx, cfg.GCSBucket, cfg.GCPProjectID)
	if err != nil {
		log.Fatalf("falha ao criar storage do GCS: %v", err)
	}
	defer storageService.Close()

	//5.2 Document AI Service
	docAIClient, err := documentai.NewClient(ctx, cfg.GCPProjectID, cfg.GCPLocation)
	if err != nil {
		log.Fatalf("falha ao criar DocAI client: %v", err)
	}
	defer docAIClient.Close()

	docExtractor := documentai.NewDocumentAIAdapter(
		*docAIClient,
		cfg.LabsProcessorID,
	)

	//6. Carregando módulos
	//6.1 Auth Service e Auth Middleware
	authService := auth.NewSupabaseAuthService()
	userRepo := supabase.NewUserRepository(dbClient)
	createUserFromIdentityUC := user.NewCreateUserFromIdentity(userRepo) //pega a identidade do supabase para autenticação.
	userHandler := handlers.NewUserHandler(createUserFromIdentityUC)
	authMiddleware := middleware.NewAuthMiddleware(authService, userRepo)
	authorizationService := authorization.NewAuthorizationService()

	//6.2 Módulo Paciente
	patientRepo := supabase.NewPatientRepository(dbClient)
	createPatientUC := patient.NewCreatePatient(patientRepo)
	getPatientUC := patient.NewGetPatient(patientRepo)
	updatePatientUC := patient.NewUpdatePatient(patientRepo, authorizationService)
	listPatientsUC := patient.NewListPatients(patientRepo)
	patientHandler := handlers.NewPatientHandler(
		createPatientUC,
		getPatientUC,
		updatePatientUC,
		listPatientsUC,
	)

	//6.3 Módulo Lab Reports
	labReportRepo := supabase.NewLabsRepository(dbClient)
	createLabReportUC := labs.NewExtractLabs(labReportRepo, docExtractor)
	listLabsUC := labs.NewListLabs(patientRepo, labReportRepo)
	listFullLabsUC := labs.NewListFullLabs(patientRepo, labReportRepo)
	labReportHandler := handlers.NewLabsHandler(
		createLabReportUC,
		listLabsUC,
		listFullLabsUC,
		storageService,
	)

	// 7.Configura o Gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog(appLogger))

	// montar gin e rotas
	http.SetupRoutes(r, authMiddleware, userHandler, patientHandler, labReportHandler)

	slog.Info("API running", "url", "http://localhost:"+cfg.Port+"/api/v1")
	if err := r.Run(":" + cfg.Port); err != nil {
		// 1. Loga o erro com nível Error (estruturado)
		slog.Error("failed to start server", "error", err)

		// 2. Encerra o programa manualmente com código de erro 1
		os.Exit(1)
	}

}
