// internal/api/app.go
package api

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/humaerror"
	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gin-gonic/gin"
)

type Options struct {
	Name       string
	Version    string
	Logger     *slog.Logger
	Deps       *APIDependencies
	CORSConfig config.CORSConfig
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

	// Assigned during route setup before the server accepts requests. Keeping
	// recovery registered here also protects Huma's docs and native Gin routes.
	var humaAPI huma.API
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(logger),
		middleware.Recovery(logger, func(c *gin.Context, err error) {
			op := &huma.Operation{Method: c.Request.Method, Path: c.FullPath()}
			_ = humaerror.Write(humaAPI, humagin.NewContext(op, c), err)
		}),
		// Keep CORS inside observability/recovery so even early responses are traced.
		middleware.SetupCors(opts.CORSConfig),
	)

	deps := *opts.Deps
	deps.APIInfo = APIInfo{
		Name:    opts.Name,
		Version: opts.Version,
	}
	humaAPI = SetupRoutes(r, &deps)

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
