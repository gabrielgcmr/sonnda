// internal/adapters/inbound/http/router.go
package httpserver

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/api"
	sharedmw "sonnda-api/internal/adapters/inbound/http/shared/middleware"
	"sonnda-api/internal/adapters/inbound/http/web"
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

	// 1. Criando um sub-sistema de arquivos para a pasta 'public'
	// Isso remove a necessidade de ter a pasta física no servidor de produção
	publicFiles, err := fs.Sub(web.PublicFS, "public")
	if err != nil {
		panic("Falha ao carregar assets embutidos: " + err.Error())
	}

	// 2. Servindo os arquivos estáticos via rota /assets
	// Agora ele lê do binário, não do disco
	r.StaticFS("/assets", http.FS(publicFiles))

	// 3. Favicon Simplificado
	// Como o favicon está dentro de public/, você pode apenas redirecionar ou servir direto
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", http.FS(publicFiles))
	})

	// ---- Rotas ----
	// Rotas
	web.SetupRoutes(r, deps.WEB)
	api.SetupRoutes(r, deps.API)

	return r
}
