// internal/api/helpers/current_user.go
package helpers

import (
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"

	"github.com/gin-gonic/gin"
)

const CurrentUserKey = "current_user"

func SetCurrentUser(c *gin.Context, u *accountdomain.User) { c.Set(CurrentUserKey, u) }

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
