// internal/api/huma.go
package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	accounthttp "github.com/gabrielgcmr/sonnda/internal/features/account/http"
	authhttp "github.com/gabrielgcmr/sonnda/internal/features/auth/http"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

const bearerAuthScheme = "bearerAuth"

type healthResponse struct {
	Status string `json:"status" doc:"Status da API" example:"ok"`
}

type healthOutput struct {
	Body healthResponse
}

func newHumaAPI(r *gin.Engine) huma.API {
	config := huma.DefaultConfig("Sonnda API", "dev")
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		bearerAuthScheme: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	return humagin.New(r, config)
}

func registerHumaRoutes(api huma.API, deps *APIDependencies) {
	registerHumaPublicRoutes(api)

	authenticated := huma.NewGroup(api)
	authenticated.UseMiddleware(requireBearer(api, deps.Auth))

	registered := huma.NewGroup(authenticated)
	registered.UseMiddleware(requireRegisteredAccount(api, deps.Account))

	deps.AccountHandler.RegisterHumaRoutes(authenticated, registered, bearerSecurity())
	deps.PatientAccessHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.PatientCreationHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.PatientHandler.RegisterHumaRoutes(registered, bearerSecurity())
}

func registerHumaPublicRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getHealth",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Summary:     "Health check",
		Tags:        []string{"Health"},
	}, func(_ context.Context, _ *struct{}) (*healthOutput, error) {
		return &healthOutput{Body: healthResponse{Status: "ok"}}, nil
	})
}

func bearerSecurity() []map[string][]string {
	return []map[string][]string{{bearerAuthScheme: {}}}
}

func requireBearer(api huma.API, auth *authhttp.Middleware) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		identity, err := auth.AuthenticateBearer(ctx.Context(), ctx.Header("Authorization"))
		if err != nil {
			writeHumaError(api, ctx, err)
			return
		}

		ginContext := humagin.Unwrap(ctx)
		ginContext.Request = ginContext.Request.WithContext(authhttp.ContextWithIdentity(ctx.Context(), identity))
		next(ctx)
	}
}

func requireRegisteredAccount(api huma.API, account *accounthttp.Middleware) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		identity, ok := authhttp.GetIdentityFromContext(ctx.Context())
		if !ok {
			writeHumaError(api, ctx, apperr.Unauthorized("autenticação necessária"))
			return
		}

		currentUser, err := account.ResolveRegisteredUser(ctx.Context(), identity)
		if err != nil {
			writeHumaError(api, ctx, err)
			return
		}

		ginContext := humagin.Unwrap(ctx)
		ginContext.Request = ginContext.Request.WithContext(helpers.ContextWithCurrentUser(ctx.Context(), currentUser))
		next(ctx)
	}
}

func writeHumaError(api huma.API, ctx huma.Context, err error) {
	status, problem := presenter.ToProblem(err, presenter.ProblemMeta{})
	huma.WriteErr(api, ctx, status, problem.Detail)
}
