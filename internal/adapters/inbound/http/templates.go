// internal/adapters/inbound/http/templates.go
package httpserver

import (
	"embed"
	"html/template"
)

// Embedding inbute os templates no binário final do Go.
//
//go:embed web/templates/**/*.html
var viewsFS embed.FS

func mustLoadTemplates() *template.Template {
	// ParseFS aceita patterns glob dentro do FS embutido
	t, err := template.New("").ParseFS(
		viewsFS,
		"web/templates/layouts/*.html",
		"web/templates/pages/*.html",
		"web/templates/partials/*.html",
	)
	if err != nil {
		panic(err)
	}
	return t
}
