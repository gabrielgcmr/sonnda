package helpers

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"

	"github.com/gin-gonic/gin"
)

const CurrentUserKey = "current_user"

func SetCurrentUser(c *gin.Context, u *user.User) { c.Set(CurrentUserKey, u) }

func GetCurrentUser(c *gin.Context) (*user.User, bool) {
	v, ok := c.Get(CurrentUserKey)
	if !ok || v == nil {
		return nil, false
	}
	u, ok := v.(*user.User)
	return u, ok
}

func MustGetCurrentUser(c *gin.Context) *user.User {
	u, ok := GetCurrentUser(c)
	if !ok || u == nil {
		panic("current user missing in context (middleware not applied?)")
	}
	return u
}
