package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) Home(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/home", gin.H{
		"Title": "Sonnda",
	})
}

func (h *HomeHandler) CounterPartial(c *gin.Context) {
	c.HTML(http.StatusOK, "partials/counter", gin.H{
		"Count": 42,
	})
}
