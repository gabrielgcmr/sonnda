// internal/adapters/inbound/http/templates.go
package httpserver

import (
	"embed"
	"html/template"
)

// ATENÇÃO!!! REMOVER ESSE ARQUIVO DEPOIS DE MIGRAR PARA TEMPL

// Embedding inbute os templates no binário final do Go.
//
//go:embed web/assets/templates/**/*.html
var viewsFS embed.FS

func mustLoadTemplates() *template.Template {
	// ParseFS aceita patterns glob dentro do FS embutido
	t, err := template.New("").ParseFS(
		viewsFS,
		"web/assets/templates/layouts/*.html",
		"web/assets/templates/pages/*.html",
		"web/assets/templates/partials/*.html",
	)
	if err != nil {
		panic(err)
	}
	return t
}
