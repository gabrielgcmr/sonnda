// internal/adapters/inbound/http/router.go
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api"
	sharedmw "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/middleware"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web"
	"github.com/gabrielgcmr/sonnda/internal/app/config"
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

	env := "dev"
	if infra.Config != nil && infra.Config.Env != "" {
		env = infra.Config.Env
	}

	// 1. Criando um sub-sistema de arquivos para a pasta 'static'
	// Isso remove a necessidade de ter a pasta física no servidor de produção
	bundle, err := web.LoadFS(env)
	if err != nil {
		panic("erro ao carregar filesystem: " + err.Error())
	}

	// 2. Servindo os arquivos estáticos via rota /static
	// Agora ele lê do binário, não do disco
	r.StaticFS("/static", http.FS(bundle.Static))

	// 3. Favicon Simplificado
	// Como o favicon está dentro de static/, você pode apenas redirecionar ou servir direto
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", http.FS(bundle.Static))
	})

	// ---- Rotas ----
	// Rotas
	web.SetupRoutes(r, deps.WEB)
	api.SetupRoutes(r, deps.API)

	return r
}
