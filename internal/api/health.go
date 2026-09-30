// internal/api/health.go
package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type healthResponse struct {
	Status string `json:"status" doc:"Status da API" example:"ok"`
}

type healthOutput struct {
	Body healthResponse
}

func registerHealthRoute(api huma.API) {
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
