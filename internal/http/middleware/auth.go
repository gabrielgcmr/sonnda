// internal/adapters/inbound/http/middleware/auth.go
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"sonnda-api/internal/domain/entities/identity"
	"sonnda-api/internal/domain/entities/rbac"
	"sonnda-api/internal/domain/ports/integrations"

	"github.com/gin-gonic/gin"
)

const (
	identityKey = "identity"
	// IdentityKey é a chave pública usada no gin.Context para armazenar identity.Identity.
	IdentityKey = identityKey
)

type AuthMiddleware struct {
	identityService integrations.IdentityService
}

func NewAuthMiddleware(
	identityService integrations.IdentityService,
) *AuthMiddleware {
	return &AuthMiddleware{
		identityService: identityService,
	}
}

// Authenticate validates token and loads the user from the app database.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := m.extractToken(ctx)
		if err != nil {
			m.abortUnauthorized(ctx, err.Error())
			return
		}

		id, err := m.identityService.VerifyToken(ctx.Request.Context(), token)
		if err != nil {
			m.abortUnauthorized(ctx, "token invalido ou expirado")
			return
		}

		//Seta identidade no contexto
		ctx.Set(identityKey, id)
		ctx.Next()
	}
}

func (m *AuthMiddleware) extractToken(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header ausente. use 'Bearer <idToken>' do firebase")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("formato do Authorization deve ser 'Bearer <idToken>'")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func (m *AuthMiddleware) abortUnauthorized(ctx *gin.Context, msg string) {
	ctx.Set("error_code", "unauthorized")
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":   "unauthorized",
		"message": msg,
	})
}

// Helper para obter a identidade do contexto
func GetIdentity(c *gin.Context) (*identity.Identity, bool) {
	val, ok := c.Get(identityKey)
	if !ok {
		return nil, false
	}
	id, ok := val.(*identity.Identity)
	return id, ok
}

// Helper que falha se não houver identidade
func RequireIdentity(c *gin.Context) (*identity.Identity, bool) {
	id, ok := GetIdentity(c)
	if !ok {
		c.Set("error_code", "missing_identity")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "autenticação necessária",
		})
		return nil, false
	}
	return id, true
}

func ActorFromCurrentUser(c *gin.Context) (userID string, role rbac.Role, ok bool) {
	u, ok := GetCurrentUser(c)
	if !ok || u == nil {
		return "", rbac.Role(""), false
	}

	return u.ID.String(), rbac.Role(u.Role[0]), true
}
