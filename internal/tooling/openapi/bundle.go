// internal/tooling/openapi/bundle.go
package openapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/goccy/go-yaml"
)

// Bundle validates the source and resolves external references for standalone use.
func Bundle(source string) ([]byte, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = openapi3.URIMapCache(openapi3.ReadFromFile)
	doc, err := loader.LoadFromFile(source)
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI: %w", err)
	}
	ctx := context.Background()
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate OpenAPI: %w", err)
	}
	// Remember exported schema names before InternalizeRefs replaces root refs.
	// This keeps Go DTO names stable when their definitions move to other files.
	schemaNames := map[string]string{}
	if doc.Components != nil {
		for name, ref := range doc.Components.Schemas {
			if ref.RefPath() != nil {
				schemaNames[ref.RefPath().String()] = name
			}
		}
	}
	doc.InternalizeRefs(ctx, func(doc *openapi3.T, ref openapi3.ComponentRef) string {
		if ref.CollectionName() == "schemas" && ref.RefPath() != nil {
			if name, ok := schemaNames[ref.RefPath().String()]; ok {
				return name
			}
		}
		return openapi3.DefaultRefNameResolver(doc, ref)
	})
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal OpenAPI: %w", err)
	}
	return yaml.JSONToYAML(data)
}
