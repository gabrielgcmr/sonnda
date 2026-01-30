// internal/api/helpers/identity.go
package helpers

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

const IdentityKey = "identity"

func SetIdentity(c *gin.Context, id *entity.Identity) { c.Set(IdentityKey, id) }

func GetIdentity(c *gin.Context) (*entity.Identity, bool) {

	v, ok := c.Get(IdentityKey)
	if !ok || v == nil {
		return nil, false
	}
	id, ok := v.(*entity.Identity)
	return id, ok
}
