// internal/features/auth/http/identity.go
package authhttp

import (
	"context"

	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
)

type identityContextKey struct{}

func ContextWithIdentity(ctx context.Context, identity *authdomain.Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

func GetIdentityFromContext(ctx context.Context) (*authdomain.Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(*authdomain.Identity)
	return identity, ok && identity != nil
}
