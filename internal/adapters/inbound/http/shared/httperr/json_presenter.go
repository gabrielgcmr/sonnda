// internal/adapters/inbound/http/shared/httperr/json_presenter.go

package httperr

import (
	"github.com/gin-gonic/gin"
)

// JSONPresenter apresenta erros como JSON (para API)
type JSONPresenter struct{}

func (p JSONPresenter) Present(c *gin.Context, status int, resp ErrorResponse) {
	c.JSON(status, gin.H{"error": resp})
}

// APIErrorResponder responde erros no formato JSON
func APIErrorResponder(c *gin.Context, err error) {
	BaseErrorResponder(c, err, JSONPresenter{})
}
