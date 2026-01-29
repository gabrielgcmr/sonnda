// internal/adapters/outbound/auth/supabase_bearer_provider.go
package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

type SupabaseBearerConfig struct {
	SupabaseURL string
	Issuer      string
	Audience    string
}

type SupabaseBearerProvider struct {
	verifier *oidc.IDTokenVerifier
	issuer   string
	audience string
}

var _ security.IdentityProvider = (*SupabaseBearerProvider)(nil)

func NewSupabaseBearerProvider(cfg SupabaseBearerConfig) (*SupabaseBearerProvider, error) {
	issuer := strings.TrimSpace(cfg.Issuer)
	if issuer == "" {
		issuer = deriveSupabaseIssuer(cfg.SupabaseURL)
	}
	if issuer == "" {
		return nil, errors.New("supabase issuer is required")
	}

	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	return &SupabaseBearerProvider{
		verifier: verifier,
		issuer:   issuer,
		audience: strings.TrimSpace(cfg.Audience),
	}, nil
}

func (p *SupabaseBearerProvider) Name() string {
	return "supabase-bearer"
}

func (p *SupabaseBearerProvider) AuthenticateBearerToken(ctx context.Context, bearerToken string) (*security.AuthResult, error) {
	if strings.TrimSpace(bearerToken) == "" {
		return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodBearer)}, nil
	}

	idToken, err := p.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodBearer)}, nil
	}

	var claims supabaseClaims
	if err := idToken.Claims(&claims); err != nil {
		return &security.AuthResult{Status: security.Error, Method: methodPtr(security.AuthMethodBearer)}, err
	}

	if p.audience != "" && !containsAudience(claims.Audience, p.audience) {
		return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodBearer)}, nil
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

	return &security.AuthResult{
		Status:   security.Authenticated,
		Identity: &identity,
		Method:   methodPtr(security.AuthMethodBearer),
		Claims:   claims.asMap(),
	}, nil
}

func (p *SupabaseBearerProvider) AuthenticateCookie(ctx context.Context, cookieValue string) (*security.AuthResult, error) {
	return &security.AuthResult{Status: security.Unauthenticated, Method: methodPtr(security.AuthMethodCookie)}, nil
}

type supabaseClaims struct {
	Email         string `json:"email"`
	EmailVerified *bool  `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Audience      any    `json:"aud"`
	Role          string `json:"role"`
}

func (c supabaseClaims) asMap() map[string]any {
	out := map[string]any{}
	if c.Email != "" {
		out["email"] = c.Email
	}
	if c.EmailVerified != nil {
		out["email_verified"] = *c.EmailVerified
	}
	if c.Name != "" {
		out["name"] = c.Name
	}
	if c.Picture != "" {
		out["picture"] = c.Picture
	}
	if c.Audience != nil {
		out["aud"] = c.Audience
	}
	if c.Role != "" {
		out["role"] = c.Role
	}
	return out
}

func deriveSupabaseIssuer(rawURL string) string {
	url := strings.TrimSpace(rawURL)
	if url == "" {
		return ""
	}
	url = strings.TrimRight(url, "/")
	return url + "/auth/v1"
}
