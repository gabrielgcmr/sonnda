package httpctx

import (
	"github.com/gabrielgcmr/sonnda/internal/shared/security"

	"github.com/gin-gonic/gin"
)

const IdentityKey = "identity"

func SetIdentity(c *gin.Context, id *security.Identity) { c.Set(IdentityKey, id) }

func GetIdentity(c *gin.Context) (*security.Identity, bool) {
	v, ok := c.Get(IdentityKey)
	if !ok || v == nil {
		return nil, false
	}
	id, ok := v.(*security.Identity)
	return id, ok
}
