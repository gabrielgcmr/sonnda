// internal/adapters/inbound/http/web/embed/favicon.go
package embed

import _ "embed"

//go:embed favicon.ico
var FaviconBytes []byte
