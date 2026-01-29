// internal/adapters/inbound/http/router.go
package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	sharedmw "github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gabrielgcmr/sonnda/internal/config"
)

type Infra struct {
	Logger *slog.Logger
	Config *config.Config
}

type Deps struct {
	API *APIDependencies
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
	SetupRoutes(r, deps.API)

	return r
}
