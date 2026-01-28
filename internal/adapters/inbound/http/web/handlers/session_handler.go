// internal/adapters/inbound/http/web/handlers/session_handler.go
package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/weberr"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/observability"
)

const (
	firebaseSessionCookieName = "__session"
	sessionExpiresIn          = 5 * 24 * time.Hour
)

type SessionHandler struct {
	identityService ports.IdentityService
}

func NewSessionHandler(identityService ports.IdentityService) *SessionHandler {
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
		logger := observability.FromContext(c.Request.Context())
		logger.Error("failed to bind session request",
			slog.String("error", err.Error()),
		)
		weberr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "corpo da requisição inválido, esperado: {\"id_token\": \"...\"}",
			Cause:   err,
		})
		return
	}

	if req.IDToken == "" {
		logger := observability.FromContext(c.Request.Context())
		logger.Error("id_token is empty")
		weberr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "id_token não pode estar vazio",
			Cause:   nil,
		})
		return
	}

	sessionCookie, err := h.identityService.CreateSessionCookie(c.Request.Context(), req.IDToken, sessionExpiresIn)
	if err != nil {
		code := apperr.AUTH_TOKEN_INVALID
		msg := "token invalido ou expirado"
		if isInsufficientPermissionError(err) {
			code = apperr.INFRA_AUTHENTICATION_ERROR
			msg = "falha ao criar sessao"
		}

		weberr.ErrorResponder(c, &apperr.AppError{
			Code:    code,
			Message: msg,
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

func isInsufficientPermissionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "INSUFFICIENT_PERMISSION") ||
		strings.Contains(s, "insufficient_permission") ||
		strings.Contains(s, "unexpected http response with status: 400")
}

func (h *SessionHandler) RefreshSession(c *gin.Context) {
	h.CreateSession(c)
}

func (h *SessionHandler) DeleteSession(c *gin.Context) {
	h.Logout(c)
}

func wantsRedirectAfterLogout(r *http.Request) bool {
	// Browser navigation (e.g. <form method="post">) should redirect to /login,
	// otherwise the user may end up on a blank /auth/logout page.
	if strings.EqualFold(r.Header.Get("Sec-Fetch-Mode"), "navigate") {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func (h *SessionHandler) Logout(c *gin.Context) {
	wantsRedirect := wantsRedirectAfterLogout(c.Request)
	finish := func() {
		if wantsRedirect {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		c.Status(http.StatusNoContent)
	}

	cookie, err := c.Cookie(firebaseSessionCookieName)

	// Always clear cookie (idempotent logout).
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(firebaseSessionCookieName, "", -1, "/", "", c.Request.TLS != nil, true)

	if err != nil || cookie == "" {
		finish()
		return
	}

	id, err := h.identityService.VerifySessionCookie(c.Request.Context(), cookie)
	if err != nil || id == nil {
		finish()
		return
	}

	if err := h.identityService.RevokeSessions(c.Request.Context(), id.Subject); err != nil {
		weberr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.INFRA_EXTERNAL_SERVICE_ERROR,
			Message: "nao foi possivel encerrar sessao",
			Cause:   err,
		})
		return
	}

	finish()
}

func (h *SessionHandler) GetSession(c *gin.Context) {
	cookie, err := c.Cookie(firebaseSessionCookieName)
	if err != nil || cookie == "" {
		weberr.ErrorResponder(c, &apperr.AppError{
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

		weberr.ErrorResponder(c, &apperr.AppError{
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
