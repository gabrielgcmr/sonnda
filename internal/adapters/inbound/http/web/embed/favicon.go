// internal/adapters/inbound/http/favicon.go
package httpserver

import _ "embed"

//go:embed web/assets/static/images/favicon.ico
var faviconBytes []byte
