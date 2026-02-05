package api

import (
	"log/slog"
	"net/http"

	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type Options struct {
	Name    string
	Version string
	Env     string
	Addr    string
	Logger  *slog.Logger
	Deps    *APIDependencies
}

type App struct {
	router *gin.Engine
	addr   string
}

func New(opts Options) *App {
	if opts.Addr == "" {
		opts.Addr = ":8080"
	}
	if opts.Deps == nil {
		panic("api.New: Options.Deps is required")
	}

	r := gin.New()

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Middlewares globais (infra)
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(logger),
		middleware.Recovery(logger),
	)

	registerRootRoute(r, RootInfo{
		Name:    opts.Name,
		Version: opts.Version,
		Env:     opts.Env,
	})
	SetupRoutes(r, opts.Deps)

	return &App{
		router: r,
		addr:   opts.Addr,
	}
}

func (a *App) Run() error {
	server := &http.Server{
		Addr:    a.addr,
		Handler: a.router,
	}
	return server.ListenAndServe()
}
