// internal/adapters/outbound/auth/session_provider.go
package auth

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/data"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

type SessionProvider struct {
	store data.SessionStore
}

var _ security.IdentityProvider = (*SessionProvider)(nil)

func NewSessionProvider(store data.SessionStore) *SessionProvider {
	return &SessionProvider{store: store}
}

func (p *SessionProvider) Name() string {
	return "redis-session"
}

func (p *SessionProvider) AuthenticateBearerToken(ctx context.Context, bearerToken string) (*security.AuthResult, error) {
	return &security.AuthResult{
		Status: security.Unauthenticated,
		Method: methodPtr(security.AuthMethodBearer),
	}, nil
}

func (p *SessionProvider) AuthenticateCookie(ctx context.Context, cookieValue string) (*security.AuthResult, error) {
	if cookieValue == "" {
		return &security.AuthResult{
			Status: security.Unauthenticated,
			Method: methodPtr(security.AuthMethodCookie),
		}, nil
	}

	identity, ok, err := p.store.Find(ctx, cookieValue)
	if err != nil {
		return &security.AuthResult{
			Status: security.Error,
			Method: methodPtr(security.AuthMethodCookie),
		}, err
	}
	if !ok || identity == nil {
		return &security.AuthResult{
			Status: security.Unauthenticated,
			Method: methodPtr(security.AuthMethodCookie),
		}, nil
	}

	return &security.AuthResult{
		Status:   security.Authenticated,
		Identity: identity,
		Method:   methodPtr(security.AuthMethodCookie),
	}, nil
}

func methodPtr(m security.AuthMethod) *security.AuthMethod {
	return &m
}
