// internal/features/auth/http/identity.go
package authhttp

import (
	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
	"github.com/gin-gonic/gin"
)

const IdentityKey = "identity"

func SetIdentity(c *gin.Context, id *authdomain.Identity) { c.Set(IdentityKey, id) }

func GetIdentity(c *gin.Context) (*authdomain.Identity, bool) {

	v, ok := c.Get(IdentityKey)
	if !ok || v == nil {
		return nil, false
	}
	id, ok := v.(*authdomain.Identity)
	return id, ok
}
