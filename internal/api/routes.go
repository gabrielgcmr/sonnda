// internal/api/routes.go
package api

import (
	"net/http"

	accounthttp "github.com/gabrielgcmr/sonnda/internal/features/account/http"
	authhttp "github.com/gabrielgcmr/sonnda/internal/features/auth/http"
	documentprocessinghttp "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/http"
	accesshttp "github.com/gabrielgcmr/sonnda/internal/features/patient/access/http"
	laboratoryhttp "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/http"
	patienthttp "github.com/gabrielgcmr/sonnda/internal/features/patient/http"
	profilehttp "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/http"
	"github.com/gin-gonic/gin"
)

type APIDependencies struct {
	RootInfo               RootInfo
	Auth                   *authhttp.Middleware
	Account                *accounthttp.Middleware
	AccountHandler         *accounthttp.Handler
	PatientAccessHandler   *accesshttp.Handler
	PatientCreationHandler *patienthttp.CreationHandler
	PatientHandler         *profilehttp.Handler
	LaboratoryHandler      *laboratoryhttp.Handler
	ExamsHandler           *documentprocessinghttp.ExamsHandler
}

type RootInfo struct {
	Name    string
	Version string
	Env     string
}

type RootResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Docs        string `json:"docs"`
	OpenAPI     string `json:"openapi"`
	Health      string `json:"health"`
	Ready       string `json:"ready"`
}

func SetupRoutes(r *gin.Engine, deps *APIDependencies) {
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/x-icon", faviconData)
	})

	rootInfo := normalizedRootInfo(deps.RootInfo)
	humaAPI := newHumaAPI(r, rootInfo)
	registerHumaRoutes(humaAPI, deps, rootInfo)

	// These endpoints have not moved to Huma yet, but no route path is versioned.
	registered := r.Group("")
	registered.Use(
		deps.Auth.RequireBearer(),
		deps.Account.RequireRegisteredUser())

	patients := registered.Group("/patients")
	exams := patients.Group("/:patientId/exames")
	exams.GET("", deps.ExamsHandler.ListExamDocuments)
	exams.GET("/reports", deps.ExamsHandler.ListExamDocumentTexts)
	exams.POST("", deps.ExamsHandler.UploadExamDocument)
}

func normalizedRootInfo(info RootInfo) RootInfo {
	environment := info.Env
	if environment == "" {
		environment = "dev"
	}
	name := info.Name
	if name == "" {
		name = "Sonnda API"
	}
	version := info.Version
	if version == "" {
		version = "dev"
	}

	return RootInfo{Name: name, Version: version, Env: environment}
}

func rootResponse(info RootInfo) RootResponse {
	return RootResponse{
		Name:        info.Name,
		Version:     info.Version,
		Environment: info.Env,
		Docs:        "/docs",
		OpenAPI:     "/openapi.yaml",
		Health:      "/healthz",
		Ready:       "/readyz",
	}
}
