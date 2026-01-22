// internal/adapters/inbound/http/web/handlers/auth_handler.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Login renders the login page
func (h *AuthHandler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/login", gin.H{
		"Title": "Login - Sonnda",
	})
}

// Register renders the registration page
func (h *AuthHandler) Register(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/register", gin.H{
		"Title": "Cadastro - Sonnda",
	})
}
