// internal/features/auth/http/identity.go
package authhttp

import (
	"context"

	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
	"github.com/gin-gonic/gin"
)

const IdentityKey = "identity"

type identityContextKey struct{}

func SetIdentity(c *gin.Context, id *authdomain.Identity) {
	c.Set(IdentityKey, id)
	c.Request = c.Request.WithContext(ContextWithIdentity(c.Request.Context(), id))
}

func ContextWithIdentity(ctx context.Context, identity *authdomain.Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

func GetIdentityFromContext(ctx context.Context) (*authdomain.Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(*authdomain.Identity)
	return identity, ok && identity != nil
}

func GetIdentity(c *gin.Context) (*authdomain.Identity, bool) {

	v, ok := c.Get(IdentityKey)
	if !ok || v == nil {
		return nil, false
	}
	id, ok := v.(*authdomain.Identity)
	return id, ok
}
