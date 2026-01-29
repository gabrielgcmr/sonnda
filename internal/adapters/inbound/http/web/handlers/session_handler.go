// internal/adapters/inbound/http/web/handlers/session_handler.go
package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/web/weberr"
	authinfra "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/auth"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/data"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

const (
	defaultSessionCookieName = "__session"
	defaultSessionTTL        = 7 * 24 * time.Hour
	defaultAuthStateTTL      = 10 * time.Minute
	defaultStateCookieName   = "__auth0_state"
	defaultNonceCookieName   = "__auth0_nonce"
	defaultAfterLoginURL     = "/"
)

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	SameSite http.SameSite
	TTL      time.Duration
	Secure   bool
}

type AuthFlowConfig struct {
	StateCookieName    string
	NonceCookieName    string
	StateTTL           time.Duration
	AfterLoginRedirect string
}

type SessionHandler struct {
	store      data.SessionStore
	cookie     CookieConfig
	authClient *authinfra.Authenticator
	authFlow   AuthFlowConfig
}

func NewSessionHandler(
	store data.SessionStore,
	authClient *authinfra.Authenticator,
	cookieCfg CookieConfig,
	authFlowCfg AuthFlowConfig,
) *SessionHandler {
	if strings.TrimSpace(cookieCfg.Name) == "" {
		cookieCfg.Name = defaultSessionCookieName
	}
	if strings.TrimSpace(cookieCfg.Path) == "" {
		cookieCfg.Path = "/"
	}
	if cookieCfg.SameSite == 0 {
		cookieCfg.SameSite = http.SameSiteLaxMode
	}
	if cookieCfg.TTL <= 0 {
		cookieCfg.TTL = defaultSessionTTL
	}
	if strings.TrimSpace(authFlowCfg.StateCookieName) == "" {
		authFlowCfg.StateCookieName = defaultStateCookieName
	}
	if strings.TrimSpace(authFlowCfg.NonceCookieName) == "" {
		authFlowCfg.NonceCookieName = defaultNonceCookieName
	}
	if authFlowCfg.StateTTL <= 0 {
		authFlowCfg.StateTTL = defaultAuthStateTTL
	}
	if strings.TrimSpace(authFlowCfg.AfterLoginRedirect) == "" {
		authFlowCfg.AfterLoginRedirect = defaultAfterLoginURL
	}

	return &SessionHandler{
		store:      store,
		cookie:     cookieCfg,
		authClient: authClient,
		authFlow:   authFlowCfg,
	}
}

func (h *SessionHandler) Login(c *gin.Context) {
	if h.authClient == nil {
		weberr.ErrorResponder(c, apperr.Internal("auth0 client nao configurado", nil))
		return
	}

	state, err := newRandomToken(32)
	if err != nil {
		weberr.ErrorResponder(c, apperr.Internal("falha ao iniciar login", err))
		return
	}

	nonce, err := newRandomToken(32)
	if err != nil {
		weberr.ErrorResponder(c, apperr.Internal("falha ao iniciar login", err))
		return
	}

	h.setAuthCookie(c, h.authFlow.StateCookieName, state, h.authFlow.StateTTL)
	h.setAuthCookie(c, h.authFlow.NonceCookieName, nonce, h.authFlow.StateTTL)

	url := h.authClient.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
	c.Redirect(http.StatusFound, url)
}

func (h *SessionHandler) Callback(c *gin.Context) {
	if h.authClient == nil {
		weberr.ErrorResponder(c, apperr.Internal("auth0 client nao configurado", nil))
		return
	}

	state := strings.TrimSpace(c.Query("state"))
	code := strings.TrimSpace(c.Query("code"))
	if state == "" || code == "" {
		weberr.ErrorResponder(c, apperr.Validation("parametros de callback invalidos"))
		return
	}

	storedState, err := h.authCookieValue(c, h.authFlow.StateCookieName)
	if err != nil || storedState == "" || storedState != state {
		weberr.ErrorResponder(c, apperr.Unauthorized("estado de autenticacao invalido"))
		return
	}

	nonce, _ := h.authCookieValue(c, h.authFlow.NonceCookieName)
	defer func() {
		h.clearAuthCookie(c, h.authFlow.StateCookieName)
		h.clearAuthCookie(c, h.authFlow.NonceCookieName)
	}()

	token, err := h.authClient.Exchange(c.Request.Context(), code)
	if err != nil {
		weberr.ErrorResponder(c, apperr.Internal("falha ao trocar code por token", err))
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(rawIDToken) == "" {
		weberr.ErrorResponder(c, apperr.Internal("id_token ausente", nil))
		return
	}

	verifier := h.authClient.Verifier(&oidc.Config{ClientID: h.authClient.ClientID})
	idToken, err := verifier.Verify(c.Request.Context(), rawIDToken)
	if err != nil {
		weberr.ErrorResponder(c, apperr.Unauthorized("id_token invalido"))
		return
	}

	var claims idTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		weberr.ErrorResponder(c, apperr.Internal("falha ao ler claims", err))
		return
	}

	if nonce == "" || nonce != claims.Nonce {
		weberr.ErrorResponder(c, apperr.Unauthorized("nonce invalido"))
		return
	}

	identity := security.Identity{
		Issuer:  idToken.Issuer,
		Subject: idToken.Subject,
	}
	if claims.Email != "" {
		identity.Email = stringPtr(claims.Email)
	}
	if claims.EmailVerified != nil {
		identity.EmailVerified = claims.EmailVerified
	}
	if claims.Name != "" {
		identity.Name = stringPtr(claims.Name)
	}
	if claims.Picture != "" {
		identity.PictureURL = stringPtr(claims.Picture)
	}

	if _, err := h.createSession(c, identity); err != nil {
		weberr.ErrorResponder(c, apperr.Internal("falha ao criar sessao", err))
		return
	}

	c.Redirect(http.StatusSeeOther, h.authFlow.AfterLoginRedirect)
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

	sessionID, err := h.sessionIDFromCookie(c)

	// Always clear cookie (idempotent logout).
	h.clearSessionCookie(c)

	if err != nil || sessionID == "" {
		finish()
		return
	}

	if err := h.store.Delete(c.Request.Context(), sessionID); err != nil {
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
	sessionID, err := h.sessionIDFromCookie(c)
	if err != nil || sessionID == "" {
		weberr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "nenhuma sessao ativa",
			Cause:   err,
		})
		return
	}

	identity, ok, err := h.store.Find(c.Request.Context(), sessionID)
	if err != nil {
		weberr.ErrorResponder(c, &apperr.AppError{
			Code:    apperr.INFRA_EXTERNAL_SERVICE_ERROR,
			Message: "falha ao validar sessao",
			Cause:   err,
		})
		return
	}
	if !ok || identity == nil {
		h.clearSessionCookie(c)
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
			"issuer":  identity.Issuer,
			"subject": identity.Subject,
			"email":   identity.Email,
		},
	})
}

func (h *SessionHandler) createSession(ctx *gin.Context, identity security.Identity) (string, error) {
	sessionID, err := newRandomToken(32)
	if err != nil {
		return "", err
	}

	if err := h.store.Save(ctx.Request.Context(), sessionID, identity, h.cookie.TTL); err != nil {
		return "", err
	}

	h.setSessionCookie(ctx, sessionID)
	return sessionID, nil
}

func (h *SessionHandler) sessionIDFromCookie(c *gin.Context) (string, error) {
	cookie, err := c.Cookie(h.cookie.Name)
	if err != nil || strings.TrimSpace(cookie) == "" {
		return "", err
	}
	return strings.TrimSpace(cookie), nil
}

func (h *SessionHandler) setSessionCookie(c *gin.Context, sessionID string) {
	secure := h.cookie.Secure || c.Request.TLS != nil
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(
		h.cookie.Name,
		sessionID,
		int(h.cookie.TTL.Seconds()),
		h.cookie.Path,
		h.cookie.Domain,
		secure,
		true,
	)
}

func (h *SessionHandler) clearSessionCookie(c *gin.Context) {
	secure := h.cookie.Secure || c.Request.TLS != nil
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(
		h.cookie.Name,
		"",
		-1,
		h.cookie.Path,
		h.cookie.Domain,
		secure,
		true,
	)
}

func (h *SessionHandler) setAuthCookie(c *gin.Context, name, value string, ttl time.Duration) {
	secure := h.cookie.Secure || c.Request.TLS != nil
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(
		name,
		value,
		int(ttl.Seconds()),
		h.cookie.Path,
		h.cookie.Domain,
		secure,
		true,
	)
}

func (h *SessionHandler) clearAuthCookie(c *gin.Context, name string) {
	secure := h.cookie.Secure || c.Request.TLS != nil
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(
		name,
		"",
		-1,
		h.cookie.Path,
		h.cookie.Domain,
		secure,
		true,
	)
}

func (h *SessionHandler) authCookieValue(c *gin.Context, name string) (string, error) {
	value, err := c.Cookie(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func wantsRedirectAfterLogout(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("Sec-Fetch-Mode"), "navigate") {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

type idTokenClaims struct {
	Email         string `json:"email"`
	EmailVerified *bool  `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Nonce         string `json:"nonce"`
}

func stringPtr(v string) *string {
	return &v
}

func newRandomToken(size int) (string, error) {
	if size <= 0 {
		size = 32
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(buf)
	if strings.TrimSpace(encoded) == "" {
		return "", errors.New("failed to generate token")
	}
	return encoded, nil
}
