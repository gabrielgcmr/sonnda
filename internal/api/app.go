// internal/api/app.go
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
	Logger  *slog.Logger
	Deps    *APIDependencies
}

type App struct {
	router *gin.Engine
}

func New(opts Options) *App {
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
	}
}

func (a *App) Run(addr string) error {
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{
		Addr:    addr,
		Handler: a.router,
	}
	return server.ListenAndServe()
}
