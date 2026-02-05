// internal/api/openapi/embed.go
package openapi

import _ "embed"

//go:embed openapi.yaml
var OpenAPISpec []byte
