// internal/api/openapi.go
package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

// OpenAPI creates the Huma specification without connecting to infrastructure.
// Route registration only captures operation metadata; handlers are never run.
func OpenAPI(info RootInfo) *huma.OpenAPI {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rootInfo := normalizedRootInfo(info)
	humaAPI := newHumaAPI(router, rootInfo)
	registerHumaRoutes(humaAPI, &APIDependencies{})
	return humaAPI.OpenAPI()
}

const bearerAuthScheme = "bearerAuth"

type RootInfo struct {
	Name    string
	Version string
	Env     string
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

func newHumaAPI(r *gin.Engine, info RootInfo) huma.API {
	config := huma.DefaultConfig(info.Name, info.Version)
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		bearerAuthScheme: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	return humagin.New(r, config)
}
