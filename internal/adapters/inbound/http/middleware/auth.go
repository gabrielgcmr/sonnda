// internal/adapters/inbound/http/middleware/auth.go
package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
	"sonnda-api/internal/core/ports/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	identityKey    = "identity"
	currentUserKey = "current_user"
)

type AuthMiddleware struct {
	authService services.AuthService
	userRepo    repositories.UserRepository
}

func NewAuthMiddleware(
	authService services.AuthService,
	userRepo repositories.UserRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		userRepo:    userRepo,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. Extrai token do header
		token, err := m.extractToken(ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			ctx.Abort()
			return
		}

		// 2. Usa o AuthService (que hoje é o SupabaseAuthService)
		identity, err := m.authService.VerifyToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "token inválido ou expirado",
			})
			ctx.Abort()
			return
		}
		ctx.Set(identityKey, identity)

		// 3. Carrega/cria o User da aplicação (tabela app_users) a partir da identidade do Supabase
		user, err := m.userRepo.FindByAuthIdentity(
			ctx.Request.Context(),
			identity.Provider,
			identity.Subject,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "não foi possível buscar o usuário",
				"details": err.Error(),
			})
			ctx.Abort()
			return
		}

		if user == nil {
			// tenta reaproveitar por email, caso o subject tenha mudado
			existingByEmail, err := m.userRepo.FindByEmail(ctx.Request.Context(), identity.Email)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_error",
					"message": "não foi possível buscar o usuário por email",
					"details": err.Error(),
				})
				ctx.Abort()
				return
			}

			if existingByEmail != nil {
				user, err = m.userRepo.UpdateAuthIdentity(
					ctx.Request.Context(),
					existingByEmail.ID,
					identity.Provider,
					identity.Subject,
				)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"error":   "internal_error",
						"message": "não foi possível atualizar o usuário",
						"details": err.Error(),
					})
					ctx.Abort()
					return
				}
			} else {
				user = &domain.User{
					ID:           uuid.New(),
					AuthProvider: identity.Provider,
					AuthSubject:  identity.Subject,
					Email:        identity.Email,
					Role:         domain.RolePatient,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := m.userRepo.Create(ctx.Request.Context(), user); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"error":   "internal_error",
						"message": "não foi possível criar o usuário",
						"details": err.Error(),
					})
					ctx.Abort()
					return
				}
			}
		}

		ctx.Set(currentUserKey, user)
		ctx.Next()
	}
}

// extractToken extrai o token do header Authorization
func (m *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("header Authorization não encontrado")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("formato do Authorization deve ser 'Bearer <token>'")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func Identity(c *gin.Context) (*services.Identity, bool) {
	val, ok := c.Get(identityKey)
	if !ok {
		return nil, false
	}
	id, ok := val.(*services.Identity)
	return id, ok
}

func CurrentUser(c *gin.Context) (*domain.User, bool) {
	user, exists := c.Get(currentUserKey)
	if !exists {
		return nil, false
	}

	u, ok := user.(*domain.User)
	return u, ok
}

func RequireUser(c *gin.Context, log *slog.Logger) (*domain.User, bool) {
	u, ok := CurrentUser(c)
	if !ok || u == nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return nil, false
	}
	return u, true
}
