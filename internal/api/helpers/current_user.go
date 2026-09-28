// internal/api/helpers/current_user.go
package helpers

import (
	"context"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"

	"github.com/gin-gonic/gin"
)

const CurrentUserKey = "current_user"

type currentUserContextKey struct{}

func SetCurrentUser(c *gin.Context, u *accountdomain.User) {
	c.Set(CurrentUserKey, u)
	c.Request = c.Request.WithContext(ContextWithCurrentUser(c.Request.Context(), u))
}

func ContextWithCurrentUser(ctx context.Context, user *accountdomain.User) context.Context {
	return context.WithValue(ctx, currentUserContextKey{}, user)
}

// GetCurrentUserFromContext makes the authenticated account available to HTTP
// adapters, such as Huma, that expose only the standard request context.
func GetCurrentUserFromContext(ctx context.Context) (*accountdomain.User, bool) {
	u, ok := ctx.Value(currentUserContextKey{}).(*accountdomain.User)
	return u, ok && u != nil
}

func GetCurrentUser(c *gin.Context) (*accountdomain.User, bool) {
	v, ok := c.Get(CurrentUserKey)
	if !ok || v == nil {
		return nil, false
	}
	u, ok := v.(*accountdomain.User)
	return u, ok
}

func MustGetCurrentUser(c *gin.Context) *accountdomain.User {
	u, ok := GetCurrentUser(c)
	if !ok || u == nil {
		panic("current user missing in context (middleware not applied?)")
	}
	return u
}
