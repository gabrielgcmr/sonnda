package reqctx

import (
	"sonnda-api/internal/domain/model/user"

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
