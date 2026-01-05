// internal/http/middleware/auth.go
package middleware

import (
	"errors"
	"strings"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/app/interfaces/external"
	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/model/user"
	httperrors "sonnda-api/internal/http/errors"

	"github.com/gin-gonic/gin"
)

const (
	identityKey = "identity"
	IdentityKey = identityKey
)

var (
	errAuthorizationHeaderMissing = errors.New("authorization header missing")
	errAuthorizationHeaderInvalid = errors.New("invalid authorization header format")
)

type AuthMiddleware struct {
	identityService external.IdentityService
}

func NewAuthMiddleware(identityService external.IdentityService) *AuthMiddleware {
	return &AuthMiddleware{identityService: identityService}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := extractBearerToken(ctx)
		if err != nil {
			switch {
			case errors.Is(err, errAuthorizationHeaderMissing):
				m.abortUnauthorized(ctx, apperr.AUTH_REQUIRED, "autenticação necessária", err)
			default:
				m.abortUnauthorized(ctx, apperr.AUTH_TOKEN_INVALID, "token inválido", err)
			}
			return
		}

		id, err := m.identityService.VerifyToken(ctx.Request.Context(), token)
		if err != nil {
			m.abortUnauthorized(ctx, apperr.AUTH_TOKEN_INVALID, "token inválido ou expirado", err)
			return
		}

		ctx.Set(identityKey, id)
		ctx.Next()
	}
}

func extractBearerToken(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", errAuthorizationHeaderMissing
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errAuthorizationHeaderInvalid
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func (m *AuthMiddleware) abortUnauthorized(ctx *gin.Context, code apperr.ErrorCode, msg string, cause error) {
	httperrors.WriteError(ctx, &apperr.AppError{
		Code:    code,
		Message: msg,
		Cause:   cause,
	})
	ctx.Abort()
}

func GetIdentity(c *gin.Context) (*identity.Identity, bool) {
	val, ok := c.Get(identityKey)
	if !ok {
		return nil, false
	}
	id, ok := val.(*identity.Identity)
	return id, ok
}

func RequireIdentity(c *gin.Context) (*identity.Identity, bool) {
	id, ok := GetIdentity(c)
	if !ok {
		httperrors.WriteError(c, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		})
		c.Abort()
		return nil, false
	}
	return id, true
}

func ActorFromCurrentUser(c *gin.Context) (userID string, at user.AccountType, ok bool) {
	u, ok := GetCurrentUser(c)
	if !ok || u == nil {
		return "", "", false
	}
	return u.ID.String(), u.AccountType, true
}
