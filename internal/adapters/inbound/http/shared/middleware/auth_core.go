// internal/adapters/inbound/http/middleware/auth.go
package middleware

import (
	"errors"
	"strings"

	httperr "sonnda-api/internal/adapters/inbound/http/api/httperr"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/model/user"
	external "sonnda-api/internal/domain/ports/integration"

	"github.com/gin-gonic/gin"
)

const (
	IdentityKey       = "identity"
	sessionCookieName = "__session"
)

var (
	errAuthorizationHeaderMissing = errors.New("authorization header missing")
	errAuthorizationHeaderInvalid = errors.New("invalid authorization header format")
	errSessionCookieMissing       = errors.New("session cookie missing")
)

type AuthCore struct {
	identityService external.IdentityService
}

func NewAuthCore(identityService external.IdentityService) *AuthCore {
	return &AuthCore{identityService: identityService}
}

func (m *AuthCore) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Try to authenticate using a Firebase session cookie first (web UI flow).
		if cookie, err := extractSessionCookieFromCookie(ctx); err == nil {
			id, err := m.identityService.VerifySessionCookie(ctx.Request.Context(), cookie)
			if err != nil {
				m.abortUnauthorized(ctx, apperr.AUTH_TOKEN_INVALID, "token inválido ou expirado", err)
				return
			}

			ctx.Set(IdentityKey, id)
			ctx.Next()
			return
		}

		// Fallback to Authorization header (API/mobile flow).
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

		ctx.Set(IdentityKey, id)
		ctx.Next()
	}
}

func extractSessionCookieFromCookie(ctx *gin.Context) (string, error) {
	cookie, err := ctx.Cookie(sessionCookieName)
	if err != nil || cookie == "" {
		return "", errSessionCookieMissing
	}
	return cookie, nil
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

func (m *AuthCore) abortUnauthorized(ctx *gin.Context, code apperr.ErrorCode, msg string, cause error) {
	httperr.WriteError(ctx, &apperr.AppError{
		Code:    code,
		Message: msg,
		Cause:   cause,
	})
	ctx.Abort()
}

func GetIdentity(c *gin.Context) (*identity.Identity, bool) {
	val, ok := c.Get(IdentityKey)
	if !ok {
		return nil, false
	}
	id, ok := val.(*identity.Identity)
	return id, ok
}

func ActorFromCurrentUser(c *gin.Context) (userID string, at user.AccountType, ok bool) {
	u, ok := GetCurrentUser(c)
	if !ok || u == nil {
		return "", "", false
	}
	return u.ID.String(), u.AccountType, true
}
