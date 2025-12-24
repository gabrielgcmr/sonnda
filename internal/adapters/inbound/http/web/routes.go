package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static/**
var staticFS embed.FS

func SetupRoutes(r *gin.Engine) {
	r.SetHTMLTemplate(loadTemplates())
	staticRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	r.StaticFS("/static", http.FS(staticRoot))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base", gin.H{
			"Title":           "Sonnda Medical",
			"ContentTemplate": "home",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base", gin.H{
			"Title":           "Login - Sonnda",
			"ContentTemplate": "login",
		})
	})

	r.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base", gin.H{
			"Title":           "Criar conta - Sonnda",
			"ContentTemplate": "signup",
		})
	})

	r.GET("/hello", func(c *gin.Context) {
		c.HTML(http.StatusOK, "hello", gin.H{
			"Message": "Hello from HTMX",
		})
	})
}
