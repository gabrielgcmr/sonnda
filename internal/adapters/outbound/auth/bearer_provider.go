// internal/adapters/outbound/auth/bearer_provider.go
package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

type BearerProvider struct {
	verifier *oidc.IDTokenVerifier
	audience string
}

var _ security.IdentityProvider = (*BearerProvider)(nil)

func NewBearerProvider(client *Authenticator) (*BearerProvider, error) {
	if client == nil || client.Provider == nil {
		return nil, errors.New("auth0 client is required")
	}

	audience := envOrFallback("AUTH0_API_AUDIENCE", "AUTH0_AUDIENCE")
	config := &oidc.Config{ClientID: client.ClientID}
	if audience != "" {
		config.SkipClientIDCheck = true
	}

	return &BearerProvider{
		verifier: client.Verifier(config),
		audience: audience,
	}, nil
}

func (p *BearerProvider) Name() string {
	return "auth0-bearer"
}

func (p *BearerProvider) AuthenticateBearerToken(ctx context.Context, bearerToken string) (*security.AuthResult, error) {
	if strings.TrimSpace(bearerToken) == "" {
		return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodBearer)}, nil
	}

	idToken, err := p.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodBearer)}, nil
	}

	var claims bearerClaims
	if err := idToken.Claims(&claims); err != nil {
		return &security.AuthResult{Status: security.Error, Method: methodPtr(security.AuthMethodBearer)}, err
	}

	if p.audience != "" && !containsAudience(claims.Audience, p.audience) {
		return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodBearer)}, nil
	}

	identity := security.Identity{
		Issuer:  idToken.Issuer,
		Subject: idToken.Subject,
		Scopes:  normalizeScopes(claims.Scope, claims.Scopes),
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

	return &security.AuthResult{
		Status:   security.Authenticated,
		Identity: &identity,
		Method:   methodPtr(security.AuthMethodBearer),
	}, nil
}

func (p *BearerProvider) AuthenticateCookie(ctx context.Context, cookieValue string) (*security.AuthResult, error) {
	return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodCookie)}, nil
}

type bearerClaims struct {
	Email         string   `json:"email"`
	EmailVerified *bool    `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Scope         string   `json:"scope"`
	Scopes        []string `json:"scp"`
	Audience      any      `json:"aud"`
}

func normalizeScopes(scope string, scopes []string) []string {
	if len(scopes) > 0 {
		return scopes
	}
	if strings.TrimSpace(scope) == "" {
		return nil
	}
	return strings.Fields(scope)
}

func containsAudience(raw any, audience string) bool {
	for _, aud := range normalizeAudience(raw) {
		if aud == audience {
			return true
		}
	}
	return false
}

func normalizeAudience(raw any) []string {
	switch v := raw.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
