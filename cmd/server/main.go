// cmd/server/main.go
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/gabrielgcmr/sonnda/internal/application/bootstrap"
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/observability"

	"github.com/gabrielgcmr/sonnda/internal/api"
	httpserver "github.com/gabrielgcmr/sonnda/internal/api"
	apimw "github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/ai"
	authinfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/auth"
	filestorage "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/filestorage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	redisstore "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/redis"
)

func main() {
	// 1. Carrega o contexto
	ctx := context.Background()

	// 2. Carrega configuracao
	cfg, err := config.Load()
	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr != nil && len(appErr.Violations) > 0 {
			log.Printf("Erro ao carregar configuracao: %s", appErr.Message)
			for _, v := range appErr.Violations {
				log.Printf(" - %s: %s", v.Field, v.Reason)
			}
			log.Fatal("configuração inválida")
		}
		log.Fatal("Erro ao carregar configuracao: ", err)
	}

	// 3. Carrega logger
	appLogger := observability.New(observability.Config{
		Env:       cfg.Env,
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		AppName:   "github.com/gabrielgcmr/sonnda",
		AddSource: cfg.Env == "dev",
	})
	slog.SetDefault(appLogger)

	// 4. Persistence
	// 4.1 Conectar db (Supabase via pgxpool)
	dbClient, err := postgress.NewClient(postgress.SupabaseConfig(cfg.DatabaseURL))
	if err != nil {
		log.Fatalf("falha ao criar client do supabase: %v", err)
	}
	defer dbClient.Close()

	//4.2 Redis Client (para sessões e cache)
	redisClient, err := redisstore.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("falha ao conectar ao Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("Redis conectado com sucesso")

	//6. Conectando outros servicos
	//6.1 Storage Service (GCS)
	gcpOpts := buildGCPClientOptions(cfg)
	storageService, err := filestorage.NewGCSObjectStorage(ctx, cfg.GCSBucket, cfg.GCPProjectID, gcpOpts...)
	if err != nil {
		log.Fatalf("falha ao criar storage do GCS: %v", err)
	}
	defer storageService.Close()

	//6.2 Document AI Service
	docAIClient, err := ai.NewClient(ctx, cfg.GCPProjectID, cfg.GCPLocation, gcpOpts...)
	if err != nil {
		log.Fatalf("falha ao criar DocAI client: %v", err)
	}
	defer docAIClient.Close()

	docExtractor := ai.NewDocumentAIAdapter(
		*docAIClient,
		cfg.GCPExtractLabsProcessorID,
	)

	//6.3 Auth (Supabase)
	apiAuthProvider, err := authinfra.NewSupabaseBearerProvider(authinfra.SupabaseBearerConfig{
		SupabaseURL: cfg.SupabaseProjectURL,
		Issuer:      cfg.SupabaseJWTIssuer,
		Audience:    cfg.SupabaseJWTAudience,
	})
	if err != nil {
		log.Fatalf("falha ao criar supabase bearer provider: %v", err)
	}

	//7. Módulos
	modules := bootstrap.NewModules(dbClient, docExtractor, storageService)

	//8 Middlewares
	//8.1 API
	apiAuthMW := apimw.NewAuthMiddleware(apiAuthProvider.AuthenticateBearerToken)
	apiRegMW := modules.User.RegistrationMiddleware

	//10. Cria o router HTTP
	ginMode := gin.DebugMode
	if cfg.Env == "prod" {
		ginMode = gin.ReleaseMode
	}
	gin.SetMode(ginMode)

	handler := httpserver.NewRouter(
		httpserver.Infra{
			Logger: appLogger,
			Config: cfg,
		},
		httpserver.Deps{
			API: &api.APIDependencies{
				AuthMiddleware:         apiAuthMW,
				RegistrationMiddleware: apiRegMW,
				UserHandler:            modules.User.Handler,
				PatientHandler:         modules.Patient.Handler,
				LabsHandler:            modules.Labs.Handler,
			},
		},
	)

	// 10. Inicia o servidor
	localScheme := "http"
	localPort := ":" + cfg.Port
	if cfg.Env == "prod" {
		localScheme = "https"
		localPort = ""
	}
	localAPIURL := localScheme + "://localhost" + localPort + "/v1"
	publicAPIURL := "https://api.sonnda.com.br/v1"
	slog.Info(
		"Sonnda is running",
		slog.String("mode", cfg.Env),
		slog.String("listen_addr", ":"+cfg.Port),
		slog.String("local_api_url", localAPIURL),
		slog.String("public_api_url", publicAPIURL),
	)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// 1. Loga o erro com nivel Error (estruturado)
		slog.Error("failed to start server", "error", err)

		// 2. Encerra o programa manualmente com codigo de erro 1
		os.Exit(1)
	}
}

func buildGCPClientOptions(cfg *config.Config) []option.ClientOption {
	if cfg == nil {
		return nil
	}
	if cfg.GoogleApplicationCredentialsJSON != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(cfg.GoogleApplicationCredentialsJSON))}
	}
	if cfg.GoogleApplicationCredentials != "" {
		return []option.ClientOption{option.WithCredentialsFile(cfg.GoogleApplicationCredentials)}
	}
	return nil
}
