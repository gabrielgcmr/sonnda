// internal/tooling/openapi/bundle_test.go
package openapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestBundleResolvesModulesAndPreservesSchemaNames(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "openapi.yaml", `openapi: 3.0.3
info: {title: Test, version: "1"}
paths:
  /me:
    $ref: './account.yaml#/me'
components:
  schemas:
    User:
      $ref: './schemas.yaml#/User'
`)
	writeFixture(t, dir, "account.yaml", `me:
  get:
    responses:
      "200":
        description: OK
        content:
          application/json:
            schema:
              $ref: './openapi.yaml#/components/schemas/User'
`)
	writeFixture(t, dir, "schemas.yaml", "User:\n  type: object\n  properties:\n    name:\n      type: string\n")
	source := filepath.Join(dir, "openapi.yaml")
	data, err := Bundle(source)
	if err != nil {
		t.Fatal(err)
	}
	// A standalone consumer has no source files or external-reference support.
	doc, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		t.Fatalf("bundle is not standalone: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(doc.Components.Schemas) != 1 || doc.Components.Schemas["User"] == nil {
		t.Fatalf("exported schema names changed: %v", doc.Components.Schemas)
	}
	response := doc.Paths.Value("/me").Get.Responses.Status(200).Value
	if response.Content["application/json"].Schema.Ref != "#/components/schemas/User" {
		t.Fatal("operation no longer uses the exported User schema")
	}
	for _, invalidSchema := range []string{
		"Other:\n  type: object\n",      // missing JSON pointer target
		"User:\n  type: invalid-type\n", // invalid OpenAPI schema
	} {
		writeFixture(t, dir, "schemas.yaml", invalidSchema)
		if _, err := Bundle(source); err == nil {
			t.Fatalf("invalid module was accepted: %s", invalidSchema)
		}
	}
	if err := os.Remove(filepath.Join(dir, "schemas.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := Bundle(source); err == nil {
		t.Fatal("missing module was accepted")
	}
}

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("# "+name+"\n"+content), 0o600); err != nil {
		t.Fatal(err)
	}
}
