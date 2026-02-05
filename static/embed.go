// static/embed.go
package static

import "embed"

//go:embed docs.html
var DocsHTML []byte

//go:embed favicon.ico
var FaviconICO []byte

//go:embed logo/*
var LogoFS embed.FS
