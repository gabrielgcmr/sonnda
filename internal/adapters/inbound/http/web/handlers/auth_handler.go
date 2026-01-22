// internal/adapters/inbound/http/web/handlers/auth_handler.go
package handlers

import (
	"net/http"

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
	c.HTML(http.StatusOK, "pages/login", gin.H{
		"Title": "Login - Sonnda",
		"FirebaseConfig": gin.H{
			"apiKey":            h.cfg.FirebaseAPIKey,
			"authDomain":        h.cfg.FirebaseAuthDomain,
			"projectId":         h.cfg.FirebaseProjectID,
			"storageBucket":     h.cfg.FirebaseStorageBucket,
			"messagingSenderId": h.cfg.FirebaseMessagingSenderID,
			"appId":             h.cfg.FirebaseAppID,
		},
	})
}

// Register renders the registration page
func (h *AuthHandler) Register(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/register", gin.H{
		"Title": "Cadastro - Sonnda",
		"FirebaseConfig": gin.H{
			"apiKey":            h.cfg.FirebaseAPIKey,
			"authDomain":        h.cfg.FirebaseAuthDomain,
			"projectId":         h.cfg.FirebaseProjectID,
			"storageBucket":     h.cfg.FirebaseStorageBucket,
			"messagingSenderId": h.cfg.FirebaseMessagingSenderID,
			"appId":             h.cfg.FirebaseAppID,
		},
	})
}
