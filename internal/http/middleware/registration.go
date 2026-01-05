package middleware

import (
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/app/interfaces/repositories"
	applog "sonnda-api/internal/app/observability"
	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/model/user"
	httperrors "sonnda-api/internal/http/errors"

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
			httperrors.WriteError(ctx, &apperr.AppError{
				Code:    apperr.AUTH_REQUIRED,
				Message: "autenticação necessária",
			})
			ctx.Abort()
			return
		}

		currentUser, ok := m.resolveUser(ctx, id)
		if !ok {
			return
		}

		ctx.Set(CurrentUserKey, currentUser)
		ctx.Next()
	}
}

func (m *RegistrationMiddleware) RequireUnregisteredUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := GetIdentity(ctx)
		if !ok {
			httperrors.WriteError(ctx, &apperr.AppError{
				Code:    apperr.AUTH_REQUIRED,
				Message: "autenticação necessária",
			})
			ctx.Abort()
			return
		}

		existing, err := m.userRepo.FindByAuthIdentity(
			ctx.Request.Context(),
			id.Provider,
			id.Subject,
		)
		if err != nil {
			httperrors.WriteError(ctx, &apperr.AppError{
				Code:    apperr.INFRA_DATABASE_ERROR,
				Message: "falha ao verificar registro",
				Cause:   err,
			})
			ctx.Abort()
			return
		}

		if existing != nil {
			httperrors.WriteError(ctx, &apperr.AppError{
				Code:    apperr.RESOURCE_ALREADY_EXISTS,
				Message: "usuário já cadastrado",
			})
			ctx.Abort()
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

		currentUser, err := m.userRepo.FindByAuthIdentity(
			ctx.Request.Context(),
			id.Provider,
			id.Subject,
		)
		if err != nil {
			// não aborta: melhor esforço
			// aqui sim você pode logar (é infra e não haverá resposta de erro)
			applog.FromContext(ctx.Request.Context()).Warn(
				"load_current_user_failed",
				"provider", id.Provider,
				"subject", id.Subject,
				"err", err,
			)
			ctx.Next()
			return
		}

		if currentUser != nil {
			ctx.Set(CurrentUserKey, currentUser)
		}

		ctx.Next()
	}
}

func (m *RegistrationMiddleware) resolveUser(ctx *gin.Context, identity *identity.Identity) (*user.User, bool) {
	currentUser, err := m.userRepo.FindByAuthIdentity(
		ctx.Request.Context(),
		identity.Provider,
		identity.Subject,
	)
	if err != nil {
		httperrors.WriteError(ctx, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "erro ao buscar usuário",
			Cause:   err,
		})
		ctx.Abort()
		return nil, false
	}

	if currentUser == nil {
		httperrors.WriteError(ctx, &apperr.AppError{
			Code:    apperr.ACCESS_DENIED,
			Message: "usuário não registrado",
		})
		ctx.Abort()
		return nil, false
	}

	return currentUser, true
}

// GetCurrentUser obtém o usuário do contexto (não escreve resposta).
func GetCurrentUser(c *gin.Context) (*user.User, bool) {
	val, ok := c.Get(CurrentUserKey)
	if !ok {
		return nil, false
	}
	currentUser, ok := val.(*user.User)
	return currentUser, ok
}
