// internal/api/openapi.go
package api

import _ "embed"

//go:embed assets/openapi.yaml
var openapiSpec []byte

//go:embed assets/docs.html
var docsHTML []byte

//go:embed assets/redoc.standalone.js
var redocBundle []byte
