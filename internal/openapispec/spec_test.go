// internal/openapispec/spec_test.go
package openapispec

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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

func TestEveryOperationHasUniqueOperationID(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromData(OpenAPISpec)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]string)
	for path, item := range document.Paths.Map() {
		if strings.Contains(path, "{id}") {
			t.Errorf("path %s uses generic {id}; use a semantic parameter name", path)
		}
		for method, operation := range item.Operations() {
			location := strings.ToUpper(method) + " " + path
			if operation.OperationID == "" {
				t.Errorf("%s has no operationId", location)
				continue
			}
			if previous, exists := seen[operation.OperationID]; exists {
				t.Errorf("operationId %q is shared by %s and %s", operation.OperationID, previous, location)
			}
			seen[operation.OperationID] = location
		}
	}
}
