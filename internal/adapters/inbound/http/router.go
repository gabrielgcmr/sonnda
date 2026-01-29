// internal/adapters/inbound/http/router.go
package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api"
	sharedmw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/middleware"
	"github.com/gabrielgcmr/sonnda/internal/app/config"
)

type Infra struct {
	Logger *slog.Logger
	Config *config.Config
}

type Deps struct {
	API *api.APIDependencies
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

	// ---- Rotas ----
	api.SetupRoutes(r, deps.API)

	return r
}
