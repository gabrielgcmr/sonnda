// cmd/server/main.go
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/app/bootstrap"
	"github.com/gabrielgcmr/sonnda/internal/app/config"
	"github.com/gabrielgcmr/sonnda/internal/kernel/observability"

	httpserver "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api"
	apimw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/middleware"
	sharedauth "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/auth"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web"
	webhandlers "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/handlers"
	webmw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/middleware"
	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/ai"
	authinfra "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/auth"
	postgress "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres"
	redisstore "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/redis"
	filestorage "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/file"
)

func main() {
	// 1. Carrega o contexto
	ctx := context.Background()

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
	dbClient, err := postgress.NewClient(config.SupabaseConfig(*cfg))
	if err != nil {
		log.Fatalf("falha ao criar client do supabase: %v", err)
	}
	defer dbClient.Close()

	//5.2 Redis Client (para sessões e cache)
	redisClient, err := redisstore.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("falha ao conectar ao Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("Redis conectado com sucesso")

	//6. Conectando outros servicos
	//6.1 Storage Service (GCS)
	storageService, err := filestorage.NewGCSObjectStorage(ctx, cfg.GCSBucket, cfg.GCPProjectID)
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
		cfg.GCPExtractLabsProcessorID,
	)

	//6.3 Auth (Auth0)
	webAuthClient, err := authinfra.NewWebClient()
	if err != nil {
		log.Fatalf("falha ao criar auth0 web client: %v", err)
	}
	apiAuthClient, err := authinfra.NewAPIClient()
	if err != nil {
		log.Fatalf("falha ao criar auth0 api client: %v", err)
	}
	apiAuthProvider, err := authinfra.NewBearerProvider(apiAuthClient)
	if err != nil {
		log.Fatalf("falha ao criar auth0 bearer provider: %v", err)
	}

	apiAuthCore := sharedauth.NewCore(apiAuthProvider)

	sessionStore, err := redisstore.NewSessionStore(redisClient, "session:")
	if err != nil {
		log.Fatalf("falha ao criar session store: %v", err)
	}
	webSessionProvider := authinfra.NewSessionProvider(sessionStore)
	webAuthCore := sharedauth.NewCore(
		webSessionProvider,
		sharedauth.WithSessionCookieName("__session"),
	)

	//7. Módulos
	modules := bootstrap.NewModules(dbClient, apiAuthProvider, docExtractor, storageService)

	//8 Middlewares
	//8.1 API
	apiAuthMW := apimw.NewAuthMiddleware(apiAuthCore)
	apiRegMW := apimw.NewRegistrationMiddleware(modules.User.RegistrationCore)

	//8.2 Web
	webAuthMW := webmw.NewAuthMiddleware(webAuthCore)
	webRegMW := webmw.NewRegistrationMiddleware(modules.User.RegistrationCore)

	// 9. Handlers WEB
	//TODO: Fazer bootstrap dos handlers web
	homeHandler := webhandlers.NewHomeHandler(modules.Patient.Service)
	sessionHandler := webhandlers.NewSessionHandler(
		sessionStore,
		webAuthClient,
		webhandlers.CookieConfig{
			Name:     "__session",
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
			Secure:   cfg.Env == "prod",
			TTL:      7 * 24 * time.Hour,
		},
		webhandlers.AuthFlowConfig{
			StateCookieName:    "__auth0_state",
			NonceCookieName:    "__auth0_nonce",
			StateTTL:           10 * time.Minute,
			AfterLoginRedirect: "/",
		},
	)

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
			WEB: web.WebDependencies{
				HomeHandler:     homeHandler,
				SessionHandler:  sessionHandler,
				WebAuth:         webAuthMW,
				WebRegistration: webRegMW,
			},
		},
	)

	// 10. Inicia o servidor
	localScheme := "http"
	localAppHost := cfg.AppHost
	localAPIHost := cfg.APIHost
	localPort := ":" + cfg.Port
	if cfg.Env == "prod" {
		localScheme = "https"
		localPort = ""
	}
	localAppURL := localScheme + "://" + localAppHost + localPort + "/"
	localAPIURL := localScheme + "://" + localAPIHost + localPort + "/v1"
	publicAppURL := "https://app.sonnda.com.br/"
	publicAPIURL := "https://api.sonnda.com.br/v1"
	slog.Info(
		"Sonnda is running",
		slog.String("mode", cfg.Env),
		slog.String("listen_addr", ":"+cfg.Port),
		slog.String("local_app_url", localAppURL),
		slog.String("local_api_url", localAPIURL),
		slog.String("public_app_url", publicAppURL),
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
