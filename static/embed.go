// static/embed.go
// static/embed.go
package static

import "embed"

//go:embed favicon.ico
var FaviconICO []byte

//go:embed logo/*
var LogoFS embed.FS
