// internal/config/auth0.go
package config

import (
	"errors"
	"strings"
	"time"
)

type Auth0Config struct {
	Domain       string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	LogoutURL    string
	Audience     string
	Scope        string

	SessionCookieName string
	SessionTTL        time.Duration
	SigningKey        []byte

	IssuerURL string // computed
}

func (c *Auth0Config) Validate() error {
	if c.Domain == "" || c.ClientID == "" || c.CallbackURL == "" {
		return errors.New("auth0 domain, client_id and callback_url are required")
	}
	if len(c.SigningKey) < 32 {
		return errors.New("SESSION_SIGNING_KEY must be at least 32 bytes")
	}
	return nil
}

func (c *Auth0Config) Normalize() {
	d := strings.TrimSpace(c.Domain)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	c.Domain = d
	c.IssuerURL = "https://" + d + "/"
	if c.Scope == "" {
		c.Scope = "openid profile email"
	}
	if c.SessionCookieName == "" {
		c.SessionCookieName = "sonnda_session"
	}
}
