package middleware

import (
	"net/http"

	applog "sonnda-api/internal/app/observability"
	"sonnda-api/internal/domain/entities/identity"
	"sonnda-api/internal/domain/entities/user"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/gin-gonic/gin"
)

const CurrentUserKey = "current_user"

type RegistrationMiddleware struct {
	userRepo    repositories.UserRepository
	patientRepo repositories.PatientRepository
}

func NewRegistrationMiddleware(
	userRepo repositories.UserRepository,
	patientRepo repositories.PatientRepository,
) *RegistrationMiddleware {
	return &RegistrationMiddleware{
		userRepo:    userRepo,
		patientRepo: patientRepo,
	}
}

func (m *RegistrationMiddleware) RequireRegisteredUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := GetIdentity(ctx)
		if !ok {
			ctx.Set("error_code", "missing_identity")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "autenticacao necessaria",
			})
			return
		}

		user, ok := m.resolveUser(ctx, id)
		if !ok {
			return
		}

		ctx.Set(CurrentUserKey, user)
		ctx.Next()
	}
}

func (m *RegistrationMiddleware) RequireUnregisteredUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := GetIdentity(ctx)
		if !ok {
			ctx.Set("error_code", "missing_identity")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "autenticação necessária",
			})
			return
		}

		user, err := m.userRepo.FindByAuthIdentity(
			ctx.Request.Context(),
			id.Provider,
			id.Subject,
		)

		if err != nil {
			_ = ctx.Error(err)
			ctx.Set("error_code", "internal_error")
			applog.FromContext(ctx.Request.Context()).Error(
				"find_user_by_auth_identity_failed",
				"provider", id.Provider,
				"subject", id.Subject,
				"err", err,
			)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "erro ao verificar registro",
			})
			return
		}

		if user != nil {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error":   "already_registered",
				"message": "usuário já cadastrado",
			})
			return
		}

		ctx.Next()
	}
}

// LoadCurrentUser tenta carregar o usuario e setar no contexto sem bloquear o fluxo.
func (m *RegistrationMiddleware) LoadCurrentUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := GetIdentity(ctx)
		if !ok {
			ctx.Next()
			return
		}

		user, err := m.userRepo.FindByAuthIdentity(
			ctx.Request.Context(),
			id.Provider,
			id.Subject,
		)
		if err != nil {
			_ = ctx.Error(err)
			ctx.Set("error_code", "internal_error")
			applog.FromContext(ctx.Request.Context()).Error(
				"find_user_by_auth_identity_failed",
				"provider", id.Provider,
				"subject", id.Subject,
				"err", err,
			)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "erro ao buscar usuario",
			})
			return
		}

		if user != nil {
			ctx.Set(CurrentUserKey, user)
		}

		ctx.Next()
	}
}

func (m *RegistrationMiddleware) resolveUser(ctx *gin.Context, identity *identity.Identity) (*user.User, bool) {
	user, err := m.userRepo.FindByAuthIdentity(
		ctx.Request.Context(),
		identity.Provider,
		identity.Subject,
	)
	if err != nil {
		_ = ctx.Error(err)
		ctx.Set("error_code", "internal_error")
		applog.FromContext(ctx.Request.Context()).Error(
			"find_user_by_auth_identity_failed",
			"provider", identity.Provider,
			"subject", identity.Subject,
			"err", err,
		)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "erro ao buscar usuario",
		})
		return nil, false
	}

	if user == nil {
		ctx.Set("error_code", "user_not_registered")
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "usuario nao registrado",
		})
		return nil, false
	}

	return user, true
}

// Helper para obter usuario do contexto
func GetCurrentUser(c *gin.Context) (*user.User, bool) {
	val, ok := c.Get(CurrentUserKey)
	if !ok {
		return nil, false
	}
	user, ok := val.(*user.User)
	return user, ok
}

// Helper que falha se nao houver usuario
func RequireCurrentUser(c *gin.Context) (*user.User, bool) {
	user, ok := GetCurrentUser(c)
	if !ok {
		c.Set("error_code", "missing_current_user")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "usuario nao encontrado",
		})
		return nil, false
	}
	return user, true
}
