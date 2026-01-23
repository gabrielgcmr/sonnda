package reqctx

import (
	"sonnda-api/internal/domain/model/identity"

	"github.com/gin-gonic/gin"
)

const IdentityKey = "identity"

func SetIdentity(c *gin.Context, id *identity.Identity) { c.Set(IdentityKey, id) }

func GetIdentity(c *gin.Context) (*identity.Identity, bool) {
	v, ok := c.Get(IdentityKey)
	if !ok || v == nil {
		return nil, false
	}
	id, ok := v.(*identity.Identity)
	return id, ok
}
