// internal/adapters/outbound/auth/supabase_bearer_provider.go
package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/gabrielgcmr/sonnda/internal/domain/model"
)

type SupabaseBearerConfig struct {
	SupabaseURL string
	Issuer      string
	Audience    string
}

type SupabaseBearerProvider struct {
	verifier *oidc.IDTokenVerifier
	audience string
}

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
		audience: strings.TrimSpace(cfg.Audience),
	}, nil
}

func (p *SupabaseBearerProvider) AuthenticateBearerToken(ctx context.Context, bearerToken string) (*model.Identity, error) {
	if strings.TrimSpace(bearerToken) == "" {
		return nil, nil
	}

	idToken, err := p.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return nil, nil
	}

	var claims supabaseClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	if p.audience != "" && !containsAudience(claims.Audience, p.audience) {
		return nil, nil
	}

	identity := model.Identity{
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

	return &identity, nil
}

type supabaseClaims struct {
	Email         string `json:"email"`
	EmailVerified *bool  `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Audience      any    `json:"aud"`
	Role          string `json:"role"`
}

func deriveSupabaseIssuer(rawURL string) string {
	url := strings.TrimSpace(rawURL)
	if url == "" {
		return ""
	}
	url = strings.TrimRight(url, "/")
	return url + "/auth/v1"
}
