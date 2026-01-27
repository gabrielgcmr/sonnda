// internal/adapters/inbound/http/shared/middleware/host_routing_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/app/config"

	"github.com/gin-gonic/gin"
)

func TestHostRouting_AppHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppHost: "app.sonnda.com.br",
		APIHost: "api.sonnda.com.br",
	}

	r := gin.New()
	r.Use(HostRouting(cfg))
	r.GET("/", func(c *gin.Context) {
		entry, _ := c.Get(ctxEntryKey)
		if entry != EntryApp {
			t.Fatalf("expected entry=app, got %v", entry)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://app.sonnda.com.br/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestHostRouting_APIHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppHost: "app.sonnda.com.br",
		APIHost: "api.sonnda.com.br",
	}

	r := gin.New()
	r.Use(HostRouting(cfg))
	r.GET("/", func(c *gin.Context) {
		entry, _ := c.Get(ctxEntryKey)
		if entry != EntryAPI {
			t.Fatalf("expected entry=api, got %v", entry)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://api.sonnda.com.br/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestHostRouting_UnknownHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppHost: "app.sonnda.com.br",
		APIHost: "api.sonnda.com.br",
	}

	r := gin.New()
	r.Use(HostRouting(cfg))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://unknown.sonnda.com.br/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestRequireEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ctxEntryKey, EntryApp)
		c.Next()
	})
	r.Use(RequireEntry(EntryApp))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://app.sonnda.com.br/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
