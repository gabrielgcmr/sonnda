// cmd/api/main.go
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/gabrielgcmr/sonnda/internal/app/bootstrap"
	"github.com/gabrielgcmr/sonnda/internal/app/config"
	"github.com/gabrielgcmr/sonnda/internal/app/observability"

	httpserver "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api"
	apimw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/middleware"
	sharedauth "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/auth"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web"
	webhandlers "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/handlers"
	webmw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/middleware"
	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/ai"
	authinfra "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/auth"
	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres/repository/db"
	redisstore "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/redis"
	storage "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/file"
)

func main() {
	// 1. Carrega o contexto
	ctx := context.Background()

	// 2. Setup Google Cloud credentials
	if err := config.SetupGoogleCredentials(); err != nil {
		log.Printf("Aviso: Não foi possível configurar credenciais do Google Cloud: %v", err)
	}

	// 3. Carrega variaveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env nao encontrado, usando variaveis de ambiente do sistema")
	}

	// 4. Carrega configuracao
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar configuracao: ", err)
	}

	//4. Carrega logger
	appLogger := observability.New(observability.Config{
		Env:       cfg.Env,
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		AppName:   "github.com/gabrielgcmr/sonnda",
		AddSource: cfg.Env == "dev",
	})
	slog.SetDefault(appLogger)

	// 5. Persistence
	// 5.1 Conectar db (Supabase via pgxpool)
	dbClient, err := db.NewClient(config.SupabaseConfig(*cfg))
	if err != nil {
		log.Fatalf("falha ao criar client do supabase: %v", err)
	}
	defer dbClient.Close()

	//5.2 Redis Client (para sessões e cache)
	redisClient, err := redisstore.NewClient()
	if err != nil {
		log.Fatalf("falha ao conectar ao Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("Redis conectado com sucesso")

	//6. Conectando outros servicos
	//6.1 Storage Service (GCS)
	storageService, err := storage.NewGCSObjectStorage(ctx, cfg.GCSBucket, cfg.GCPProjectID)
	if err != nil {
		log.Fatalf("falha ao criar storage do GCS: %v", err)
	}
	defer storageService.Close()

	//6.2 Document AI Service
	docAIClient, err := ai.NewClient(ctx, cfg.GCPProjectID, cfg.GCPLocation)
	if err != nil {
		log.Fatalf("falha ao criar DocAI client: %v", err)
	}
	defer docAIClient.Close()

	docExtractor := ai.NewDocumentAIAdapter(
		*docAIClient,
		cfg.LabsProcessorID,
	)

	//6.3 Auth (Firebase only)
	authService, err := authinfra.NewFirebaseAuthService(ctx)
	if err != nil {
		log.Fatalf("falha ao criar auth do firebase: %v", err)
	}
	authCore := sharedauth.NewCore(authService)

	//7. Módulos
	modules := bootstrap.NewModules(dbClient, authService, docExtractor, storageService)

	//8 Middlewares
	//8.1 API
	apiAuthMW := apimw.NewAuthMiddleware(authCore)
	apiRegMW := apimw.NewRegistrationMiddleware(modules.User.RegistrationCore)

	//8.2 Web
	webAuthMW := webmw.NewAuthMiddleware(authCore)
	webRegMW := webmw.NewRegistrationMiddleware(modules.User.RegistrationCore)

	// 9. Handlers WEB
	//TODO: Fazer bootstrap dos handlers web
	homeHandler := webhandlers.NewHomeHandler(modules.Patient.Service)
	authHandler := webhandlers.NewAuthHandler(cfg)
	sessionHandler := webhandlers.NewSessionHandler(authService)

	//10. Cria o router HTTP
	gin.SetMode(gin.ReleaseMode)

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
			WEB: web.WebDependencies{
				HomeHandler:     homeHandler,
				AuthHandler:     authHandler,
				SessionHandler:  sessionHandler,
				WebAuth:         webAuthMW,
				WebRegistration: webRegMW,
			},
		},
	)

	// 10. Inicia o servidor
	slog.Info(
		"Sonnda is running",
		slog.String("Mode: env", cfg.Env),
		slog.String("Api url", "http://localhost:"+cfg.Port+"/api/v1"),
		slog.String("App url", "http://localhost:"+cfg.Port+"/"),
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
