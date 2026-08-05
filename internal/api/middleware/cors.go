// internal/api/middleware/cors.go
package middleware

import (
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCors(cfg config.CORSConfig) gin.HandlerFunc {
	allowWildcard := false
	for _, origin := range cfg.AllowOrigins {
		if strings.Contains(origin, "*") {
			allowWildcard = true
			break
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowWildcard:    allowWildcard,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
