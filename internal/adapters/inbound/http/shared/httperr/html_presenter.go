// internal/adapters/inbound/http/shared/httperr/html_presenter.go

package httperr

import (
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// HTMLPresenter apresenta erros como HTML (para Web)
type HTMLPresenter struct{}

func (p HTMLPresenter) Present(c *gin.Context, status int, resp ErrorResponse) {
	// HTMX request => fragmento HTML
	if isHTMX(c.Request) {
		c.Data(status, "text/html; charset=utf-8", []byte(renderHTMXFragment(resp)))
		return
	}

	// Navegação normal => página completa
	c.Data(status, "text/html; charset=utf-8", []byte(renderErrorPage(status, resp)))
}

// WebErrorResponder responde erros no formato HTML
func WebErrorResponder(c *gin.Context, err error) {
	BaseErrorResponder(c, err, HTMLPresenter{})
}

// Helpers privados

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

func renderHTMXFragment(resp ErrorResponse) string {
	msg := html.EscapeString(resp.Message)
	code := html.EscapeString(string(resp.Code))

	return `<div class="alert alert-error rounded-md border p-3 text-sm" role="alert">` +
		`<strong>` + code + `</strong>: ` + msg +
		`</div>`
}

func renderErrorPage(status int, resp ErrorResponse) string {
	msg := html.EscapeString(resp.Message)
	code := html.EscapeString(string(resp.Code))

	return `<!doctype html>
<html lang="pt-br">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<title>Erro ` + strconv.Itoa(status) + `</title>
	<style>
		body { font-family: system-ui, -apple-system, sans-serif; max-width: 720px; margin: 40px auto; padding: 0 16px; }
		.error-box { background: #fee; border-left: 4px solid #c00; padding: 1rem; margin: 1rem 0; }
		.error-code { color: #c00; font-weight: 600; }
	</style>
</head>
<body>
	<main>
		<h1>Ocorreu um erro</h1>
		<div class="error-box">
			<p class="error-code">` + code + ` (HTTP ` + strconv.Itoa(status) + `)</p>
			<p>` + msg + `</p>
		</div>
		<p><a href="/">← Voltar para página inicial</a></p>
	</main>
</body>
</html>`
}
