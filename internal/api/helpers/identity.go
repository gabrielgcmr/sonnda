package helpers

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/model"

	"github.com/gin-gonic/gin"
)

const IdentityKey = "identity"

func SetIdentity(c *gin.Context, id *model.Identity) { c.Set(IdentityKey, id) }

func GetIdentity(c *gin.Context) (*model.Identity, bool) {

	v, ok := c.Get(IdentityKey)
	if !ok || v == nil {
		return nil, false
	}
	id, ok := v.(*model.Identity)
	return id, ok
}
