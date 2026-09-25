// internal/api/openapi/spec_test.go
package openapi

import (
	"bytes"
	"testing"

	openapitools "github.com/gabrielgcmr/sonnda/internal/tooling/openapi"
)

func TestEmbeddedSpecMatchesModularSource(t *testing.T) {
	want, err := openapitools.Bundle("../../../openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(OpenAPISpec, want) {
		t.Fatal("embedded OpenAPI is stale; run go generate ./internal/api/openapi")
	}
}
