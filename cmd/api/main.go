// File: cmd/api/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"sonnda-api/internal/app/bootstrap"
	"sonnda-api/internal/app/config"
	"sonnda-api/internal/app/observability"

	httpserver "sonnda-api/internal/adapters/inbound/http"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	authinfra "sonnda-api/internal/adapters/outbound/integrations/auth"
	"sonnda-api/internal/adapters/outbound/integrations/documentai"
	"sonnda-api/internal/adapters/outbound/integrations/storage"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
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
	dbClient, err := db.NewClient(config.SupabaseConfig(*cfg))
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
	authMiddleware := middleware.NewAuthMiddleware(authService)

	//7. Inicializa os repositórios
	modules := bootstrap.NewModules(dbClient, authService, docExtractor, storageService)
	registrationMiddleware := modules.User.RegistrationMiddleware

	// 8. Configura o Gin
	gin.SetMode(gin.ReleaseMode)
	r := httpserver.NewRouter(
		httpserver.Infra{Logger: appLogger},
		httpserver.Dependencies{
			AuthMiddleware:         authMiddleware,
			RegistrationMiddleware: registrationMiddleware,
			UserHandler:            modules.User.Handler,
			PatientHandler:         modules.Patient.Handler,
			LabsHandler:            modules.Labs.Handler,
		},
	)

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
