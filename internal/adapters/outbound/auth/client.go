// internal/adapters/outbound/auth/client.go
package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

func NewWebClient() (*Authenticator, error) {
	domain := envOrFallback("AUTH0_WEB_DOMAIN", "AUTH0_DOMAIN")
	clientID := envOrFallback("AUTH0_WEB_CLIENT_ID", "AUTH0_CLIENT_ID")
	clientSecret := envOrFallback("AUTH0_WEB_CLIENT_SECRET", "AUTH0_CLIENT_SECRET")
	callbackURL := envOrFallback("AUTH0_WEB_CALLBACK_URL", "AUTH0_CALLBACK_URL")

	if strings.TrimSpace(domain) == "" {
		return nil, errors.New("AUTH0_WEB_DOMAIN is required")
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, errors.New("AUTH0_WEB_CLIENT_ID is required")
	}
	if strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("AUTH0_WEB_CLIENT_SECRET is required")
	}
	if strings.TrimSpace(callbackURL) == "" {
		return nil, errors.New("AUTH0_WEB_CALLBACK_URL is required")
	}

	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+domain+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

func NewAPIClient() (*Authenticator, error) {
	domain := envOrFallback("AUTH0_API_DOMAIN", "AUTH0_WEB_DOMAIN", "AUTH0_DOMAIN")
	if strings.TrimSpace(domain) == "" {
		return nil, errors.New("AUTH0_API_DOMAIN is required")
	}

	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+domain+"/",
	)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		Provider: provider,
		Config: oauth2.Config{
			ClientID: envOrFallback("AUTH0_API_CLIENT_ID", "AUTH0_WEB_CLIENT_ID", "AUTH0_CLIENT_ID"),
			Endpoint: provider.Endpoint(),
		},
	}, nil
}
