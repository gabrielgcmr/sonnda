package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"sonnda-api/internal/app/config"
	"sonnda-api/internal/app/observability"

	"sonnda-api/internal/app/modules/labs"

	"sonnda-api/internal/app/modules/user"
	"sonnda-api/internal/http/api"

	"sonnda-api/internal/http/middleware"
	authinfra "sonnda-api/internal/infrastructure/auth"
	"sonnda-api/internal/infrastructure/documentai"
	"sonnda-api/internal/infrastructure/storage"
)

func main() {
	// 1. Carrega o contexto
	ctx := context.Background()
	// 2. Carrega variaveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env nao encontrado, usando variaveis de ambiente do sistema")
	}
	// 3. Carrega configuracao
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar configuracao: ", err)
	}

	//4. Carrega logger
	appLogger := observability.New(observability.Config{
		Env:       cfg.Env,
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		AppName:   "sonnda-api",
		AddSource: cfg.Env == "dev",
	})
	slog.SetDefault(appLogger)

	// 4. Conectar db (Supabase via pgxpool)
	dbClient, err := repository.NewClient(config.SupabaseConfig(*cfg))
	if err != nil {
		log.Fatalf("falha ao criar client do supabase: %v", err)
	}
	defer dbClient.Close()

	//5. Conectando outros servicos
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

	//6 Auth (Firebase only)
	authService, err := authinfra.NewFirebaseAuthService(ctx)
	if err != nil {
		log.Fatalf("falha ao criar auth do firebase: %v", err)
	}

	//7. Inicializa os repositórios
	accessRepo := repository.NewPatientAccessRepository(dbClient)
	patientRepo := repository.NewPatientRepository(dbClient)

	//7 Modules
	patientModule := patient.New(dbClient)
	userModule := user.New(dbClient, patientRepo, accessRepo, authService)
	labsModule := labs.New(dbClient, patientRepo, docExtractor, storageService)

	authMiddleware := middleware.NewAuthMiddleware(authService)
	registrationMiddleware := userModule.RegistrationMiddleware

	// 8. Configura o Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog(appLogger))
	r.Use(middleware.Recovery(appLogger))

	// 9. Montar rotas
	api.SetupRoutes(r, authMiddleware, registrationMiddleware, userModule.Handler, patientModule.Handler, labsModule.Handler)

	// 10. Inicia o servidor
	slog.Info(
		"API is running",
		slog.String("url", "http://localhost:"+cfg.Port+"/api/v1"),
		slog.String("env", cfg.Env),
	)
	if err := r.Run(":" + cfg.Port); err != nil {
		// 1. Loga o erro com nivel Error (estruturado)
		slog.Error("failed to start server", "error", err)

		// 2. Encerra o programa manualmente com codigo de erro 1
		os.Exit(1)
	}
}
