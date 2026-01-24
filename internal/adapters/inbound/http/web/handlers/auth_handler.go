// internal/adapters/inbound/http/web/handlers/auth_handler.go
package handlers

import (
	"net/http"

	"sonnda-api/internal/adapters/inbound/http/web/assets/templates/pages/auth"
	"sonnda-api/internal/app/config"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		cfg: cfg,
	}
}

// Login renders the login page
func (h *AuthHandler) Login(c *gin.Context) {
	vm := auth.LoginViewModel{
		Title: "Login - Sonnda",
		FirebaseConfig: auth.FirebaseConfig{
			APIKey:            h.cfg.FirebaseAPIKey,
			AuthDomain:        h.cfg.FirebaseAuthDomain,
			ProjectID:         h.cfg.FirebaseProjectID,
			StorageBucket:     h.cfg.FirebaseStorageBucket,
			MessagingSenderID: h.cfg.FirebaseMessagingSenderID,
			AppID:             h.cfg.FirebaseAppID,
		},
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := auth.Login(vm).Render(c.Request.Context(), c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

// Register renders the registration page
func (h *AuthHandler) Register(c *gin.Context) {
	vm := auth.RegisterViewModel{
		Title: "Cadastro - Sonnda",
		FirebaseConfig: auth.FirebaseConfig{
			APIKey:            h.cfg.FirebaseAPIKey,
			AuthDomain:        h.cfg.FirebaseAuthDomain,
			ProjectID:         h.cfg.FirebaseProjectID,
			StorageBucket:     h.cfg.FirebaseStorageBucket,
			MessagingSenderID: h.cfg.FirebaseMessagingSenderID,
			AppID:             h.cfg.FirebaseAppID,
		},
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := auth.Register(vm).Render(c.Request.Context(), c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}
