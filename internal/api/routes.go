// internal/api/routes.go
package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	accounthttp "github.com/gabrielgcmr/sonnda/internal/features/account/http"
	authhttp "github.com/gabrielgcmr/sonnda/internal/features/auth/http"
	documentprocessinghttp "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/http"
	accesshttp "github.com/gabrielgcmr/sonnda/internal/features/patient/access/http"
	laboratoryhttp "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/http"
	patienthttp "github.com/gabrielgcmr/sonnda/internal/features/patient/http"
	profilehttp "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/http"
	"github.com/gabrielgcmr/sonnda/static"
	"github.com/gin-gonic/gin"
)

type APIDependencies struct {
	RootInfo                      RootInfo
	Auth                          *authhttp.Middleware
	Account                       *accounthttp.Middleware
	AccountHandler                *accounthttp.Handler
	PatientAccessHandler          *accesshttp.Handler
	PatientCreationHandler        *patienthttp.CreationHandler
	PatientHandler                *profilehttp.Handler
	LaboratoryHandler             *laboratoryhttp.Handler
	ExamsHandler                  *documentprocessinghttp.ExamsHandler
	TemporaryLabExtractionHandler *documentprocessinghttp.TemporaryLabExtractionHandler
}

func SetupRoutes(r *gin.Engine, deps *APIDependencies) {
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/x-icon", static.FaviconICO)
	})

	rootInfo := normalizedRootInfo(deps.RootInfo)
	humaAPI := newHumaAPI(r, rootInfo)
	registerHumaRoutes(humaAPI, deps)
}

func registerHumaRoutes(api huma.API, deps *APIDependencies) {
	registerHealthRoute(api)

	authenticated := huma.NewGroup(api)
	authenticated.UseMiddleware(middleware.RequireBearer(api, deps.Auth))

	registered := huma.NewGroup(authenticated)
	registered.UseMiddleware(middleware.RequireRegisteredAccount(api, deps.Account))

	deps.AccountHandler.RegisterHumaRoutes(authenticated, registered, bearerSecurity())
	deps.PatientAccessHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.PatientCreationHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.PatientHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.ExamsHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.TemporaryLabExtractionHandler.RegisterHumaRoutes(registered, bearerSecurity())
	deps.LaboratoryHandler.RegisterHumaRoutes(registered, bearerSecurity())
}

func bearerSecurity() []map[string][]string {
	return []map[string][]string{{bearerAuthScheme: {}}}
}
