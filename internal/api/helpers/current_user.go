// internal/api/helpers/current_user.go
package helpers

import (
	"context"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
)

type currentUserContextKey struct{}

func ContextWithCurrentUser(ctx context.Context, user *accountdomain.User) context.Context {
	return context.WithValue(ctx, currentUserContextKey{}, user)
}

// GetCurrentUserFromContext makes the authenticated account available to HTTP
// adapters, such as Huma, that expose only the standard request context.
func GetCurrentUserFromContext(ctx context.Context) (*accountdomain.User, bool) {
	u, ok := ctx.Value(currentUserContextKey{}).(*accountdomain.User)
	return u, ok && u != nil
}
