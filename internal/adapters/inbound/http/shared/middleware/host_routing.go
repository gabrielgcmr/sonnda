// internal/adapters/inbound/http/shared/middleware/host_routing.go
package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/app/config"

	"github.com/gin-gonic/gin"
)

const ctxEntryKey = "entry"

type Entry string

const (
	EntryApp Entry = "app"
	EntryAPI Entry = "api"
)

// HostRouting decide se a request é "app" ou "api" baseado no host real.
// Railway: usa somente r.Host (ignora X-Forwarded-Host).
func HostRouting(cfg *config.Config) gin.HandlerFunc {
	appHost := ""
	apiHost := ""
	if cfg != nil {
		appHost = cfg.AppHost
		apiHost = cfg.APIHost
	}

	return func(c *gin.Context) {
		if appHost == "" || apiHost == "" {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		host := normalizeHost(c.Request.Host)
		if host == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		switch host {
		case appHost:
			c.Set(ctxEntryKey, EntryApp)
		case apiHost:
			c.Set(ctxEntryKey, EntryAPI)
		default:
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}

func RequireEntry(expected Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		entryRaw, ok := c.Get(ctxEntryKey)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		entry, ok := entryRaw.(Entry)
		if !ok || entry != expected {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

func normalizeHost(h string) string {
	h = strings.TrimSpace(strings.ToLower(h))
	if h == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(h); err == nil {
		return host
	}
	return h
}
