package web

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/**/*.html
var templatesFS embed.FS

func loadTemplates() *template.Template {
	var tmpl *template.Template
	tmpl = template.New("").Funcs(template.FuncMap{
		"render": func(name string, data any) template.HTML {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
				panic(err)
			}
			return template.HTML(buf.String())
		},
	})
	tmpl = parseGlobIfExistsFS(tmpl, templatesFS, "templates/layouts/*.html")
	tmpl = parseGlobIfExistsFS(tmpl, templatesFS, "templates/pages/*.html")
	tmpl = parseGlobIfExistsFS(tmpl, templatesFS, "templates/partials/*.html")
	tmpl = parseGlobIfExistsFS(tmpl, templatesFS, "templates/components/*.html")
	return tmpl
}

func parseGlobIfExistsFS(tmpl *template.Template, fsys fs.FS, pattern string) *template.Template {
	matches, err := fs.Glob(fsys, pattern)
	if err != nil || len(matches) == 0 {
		return tmpl
	}
	parsed, err := tmpl.ParseFS(fsys, matches...)
	if err != nil {
		panic(err)
	}
	return parsed
}
