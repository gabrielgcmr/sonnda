// internal/adapters/inbound/http/web/handlers/session_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	httperr "sonnda-api/internal/adapters/inbound/http/httperr"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/ports/integration"
)

const (
	firebaseSessionCookieName = "__session"
	sessionExpiresIn          = 5 * 24 * time.Hour
)

type SessionHandler struct {
	identityService integration.IdentityService
}

func NewSessionHandler(identityService integration.IdentityService) *SessionHandler {
	return &SessionHandler{
		identityService: identityService,
	}
}

type CreateSessionRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "token invalido",
			Cause:   err,
		})
		return
	}

	sessionCookie, err := h.identityService.CreateSessionCookie(c.Request.Context(), req.IDToken, sessionExpiresIn)
	if err != nil {
		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_TOKEN_INVALID,
			Message: "token invalido ou expirado",
			Cause:   err,
		})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		firebaseSessionCookieName,
		sessionCookie,
		int(sessionExpiresIn.Seconds()),
		"/",                  // path
		"",                   // domain (empty = current domain)
		c.Request.TLS != nil, // secure (true if HTTPS)
		true,                 // httpOnly
	)

	c.Status(http.StatusNoContent)
}

func (h *SessionHandler) RefreshSession(c *gin.Context) {
	h.CreateSession(c)
}

func (h *SessionHandler) DeleteSession(c *gin.Context) {
	h.Logout(c)
}

func (h *SessionHandler) Logout(c *gin.Context) {
	cookie, err := c.Cookie(firebaseSessionCookieName)

	// Always clear cookie (idempotent logout).
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(firebaseSessionCookieName, "", -1, "/", "", c.Request.TLS != nil, true)

	if err != nil || cookie == "" {
		c.Status(http.StatusNoContent)
		return
	}

	id, err := h.identityService.VerifySessionCookie(c.Request.Context(), cookie)
	if err != nil || id == nil {
		c.Status(http.StatusNoContent)
		return
	}

	if err := h.identityService.RevokeSessions(c.Request.Context(), id.Subject); err != nil {
		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.INFRA_EXTERNAL_SERVICE_ERROR,
			Message: "nao foi possivel encerrar sessao",
			Cause:   err,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SessionHandler) GetSession(c *gin.Context) {
	cookie, err := c.Cookie(firebaseSessionCookieName)
	if err != nil || cookie == "" {
		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "nenhuma sessao ativa",
			Cause:   err,
		})
		return
	}

	id, err := h.identityService.VerifySessionCookie(c.Request.Context(), cookie)
	if err != nil {
		// Cookie exists but token is invalid/expired; clear it.
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(firebaseSessionCookieName, "", -1, "/", "", c.Request.TLS != nil, true)

		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_TOKEN_INVALID,
			Message: "sessao expirada",
			Cause:   err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"provider": id.Provider,
			"subject":  id.Subject,
			"email":    id.Email,
		},
	})
}

