// internal/openapispec/spec_test.go
package openapispec

import (
	"bytes"
	"os"
	"testing"
)

func TestEmbeddedSpecMatchesDistributionBundle(t *testing.T) {
	want, err := os.ReadFile("../../dist/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(OpenAPISpec, want) {
		t.Fatal("embedded OpenAPI is stale; run make openapi-generate")
	}
}
