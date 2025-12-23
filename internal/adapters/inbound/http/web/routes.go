package web

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.SetHTMLTemplate(loadTemplates())
	r.Static("/static", "internal/adapters/inbound/http/web/static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home", gin.H{
			"Title": "HTMX + Tailwind",
		})
	})

	r.GET("/hello", func(c *gin.Context) {
		c.HTML(http.StatusOK, "hello", gin.H{
			"Message": "Hello from HTMX",
		})
	})
}

func loadTemplates() *template.Template {
	base := "internal/adapters/inbound/http/web/templates"
	tmpl := template.New("")
	tmpl = parseGlobIfExists(tmpl, filepath.Join(base, "layouts", "*.html"))
	tmpl = parseGlobIfExists(tmpl, filepath.Join(base, "pages", "*.html"))
	tmpl = parseGlobIfExists(tmpl, filepath.Join(base, "partials", "*.html"))
	tmpl = parseGlobIfExists(tmpl, filepath.Join(base, "components", "*.html"))
	return tmpl
}

func parseGlobIfExists(tmpl *template.Template, pattern string) *template.Template {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return tmpl
	}
	return template.Must(tmpl.ParseFiles(matches...))
}
