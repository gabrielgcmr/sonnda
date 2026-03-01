// internal/api/middleware/cors.go
package middleware

import (
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCors(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
