// internal/adapters/inbound/http/router.go
package httpserver

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/api"
	sharedmw "sonnda-api/internal/adapters/inbound/http/shared/middleware"
	"sonnda-api/internal/adapters/inbound/http/web"
	"sonnda-api/internal/adapters/inbound/http/web/embed"
	"sonnda-api/internal/app/config"
)

type Infra struct {
	Logger *slog.Logger
	Config *config.Config
}

type Deps struct {
	API *api.APIDependencies
	WEB web.WebDependencies
}

func NewRouter(infra Infra, deps Deps) *gin.Engine {
	r := gin.New()

	logger := infra.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Middlewares globais (infra)
	r.Use(
		sharedmw.RequestID(),
		sharedmw.AccessLog(logger),
		sharedmw.Recovery(logger),
	)

	// Static assets (css, js, imagens)
	// Use relative path that works from project root (where air runs from)
	r.Static("/static", filepath.Join(
		"internal", "adapters", "inbound", "http", "web", "assets", "static",
	))

	// Favicon embutido no binário
	r.GET("/favicon.ico", func(c *gin.Context) {
		if len(embed.FaviconBytes) == 0 {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", embed.FaviconBytes)
	})

	// ---- Rotas ----
	// Rotas
	web.SetupRoutes(r, deps.WEB)
	api.SetupRoutes(r, deps.API)

	return r
}
