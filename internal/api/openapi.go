// internal/api/openapi.go
package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
)

// OpenAPI creates the Huma specification without connecting to infrastructure.
// Route registration only captures operation metadata; handlers are never run.
func OpenAPI(info RootInfo) *huma.OpenAPI {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rootInfo := normalizedRootInfo(info)
	humaAPI := newHumaAPI(router, rootInfo)
	registerHumaRoutes(humaAPI, &APIDependencies{}, rootInfo)
	return humaAPI.OpenAPI()
}
