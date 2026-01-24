// internal/adapters/inbound/http/web/embed.go
package web

import "embed"

//go:embed public/*
var PublicFS embed.FS
