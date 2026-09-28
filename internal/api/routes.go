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

	humaAPI := newHumaAPI(r)
	registerHumaRoutes(humaAPI, deps)
	registerPublicRoutes(r)

	// These endpoints have not moved to Huma yet, but no route path is versioned.
	registered := r.Group("")
	registered.Use(
		deps.Auth.RequireBearer(),
		deps.Account.RequireRegisteredUser())

	me := registered.Group("/me")
	me.GET("/patients", deps.PatientAccessHandler.ListForCurrentAccount)

	patients := registered.Group("/patients")
	patients.POST("", deps.PatientCreationHandler.Create)
	patients.GET("", deps.PatientHandler.ListPatients)
	patients.GET("/:patientId", deps.PatientHandler.GetPatient)
	patients.GET("/:patientId/lab-reports", deps.LaboratoryHandler.ListLabs)
	patients.GET("/:patientId/exam-documents", deps.ExamsHandler.ListExamDocuments)
	patients.POST("/:patientId/exam-documents", deps.ExamsHandler.UploadExamDocument)

	labs := patients.Group("/:patientId/labs")
	labs.GET("", deps.LaboratoryHandler.ListLabs)

	exams := patients.Group("/:patientId/exames")
	exams.GET("", deps.ExamsHandler.ListExamDocuments)
	exams.GET("/document-texts", deps.ExamsHandler.ListExamDocumentTexts)
	exams.GET("/reports", deps.ExamsHandler.ListExamDocumentTexts)
	exams.POST("", deps.ExamsHandler.UploadExamDocument)

	registered.GET("/exam-documents/:documentId", deps.ExamsHandler.GetExamDocument)
	registered.GET("/exam-documents/:documentId/file", deps.ExamsHandler.GetExamDocumentFile)
	registered.GET("/lab-reports/:labReportId", deps.LaboratoryHandler.GetLabReport)
}

func registerRootRoute(r gin.IRouter, info RootInfo) {
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

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, RootResponse{
			Name:        name,
			Version:     version,
			Environment: environment,
			Docs:        "/docs",
			OpenAPI:     "/openapi.yaml",
			Health:      "/healthz",
			Ready:       "/readyz",
		})
	})
}

func registerPublicRoutes(r gin.IRouter) {
	r.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
}
