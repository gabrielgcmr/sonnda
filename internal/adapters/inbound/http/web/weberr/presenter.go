// internal/adapters/inbound/http/shared/httperr/presenter.go
package weberr

import (
	"html"
	"net/http"
	"sonnda-api/internal/adapters/inbound/http/shared/httperr"
	"sonnda-api/internal/app/apperr"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ErrorResponder(c *gin.Context, err error) {
	if c.Writer.Written() {
		c.Abort()
		return
	}
	if err != nil {
		_ = c.Error(err)
	}

	status, resp := httperr.ToHTTP(err)

	// Copiar política do API responder:
	level := apperr.LogLevelOf(err)
	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", level)

	// HTMX request => devolve fragmento HTML (pra swap)
	if isHTMX(c.Request) {
		// opcional: forçar o swap cair num container global tipo #flash
		// c.Header("HX-Retarget", "#flash")
		// c.Header("HX-Reswap", "innerHTML")

		c.Abort()
		c.Data(status, "text/html; charset=utf-8", []byte(renderHTMXErrorFragment(status, resp)))
		return
	}

	// Navegação normal => página HTML inteira
	c.Abort()
	c.Data(status, "text/html; charset=utf-8", []byte(renderErrorPage(status, resp)))
}

func isHTMX(r *http.Request) bool {
	// HTMX envia HX-Request: true
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

func renderHTMXErrorFragment(status int, resp httperr.ErrorResponse) string {
	msg := html.EscapeString(resp.Message)
	code := html.EscapeString(string(resp.Code))
	return `<div class="rounded-md border p-3 text-sm">` +
		`<strong>` + code + `</strong>: ` + msg +
		`</div>`
}

func renderErrorPage(status int, resp httperr.ErrorResponse) string {
	msg := html.EscapeString(resp.Message)
	code := html.EscapeString(string(resp.Code))
	return `<!doctype html><html lang="pt-br"><head>` +
		`<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">` +
		`<title>Erro</title></head><body>` +
		`<main style="font-family:system-ui,Segoe UI,Roboto,Arial;max-width:720px;margin:40px auto;padding:0 16px;">` +
		`<h1>Ocorreu um erro</h1>` +
		`<p><b>` + code + `</b> (HTTP ` + itoa(status) + `)</p>` +
		`<p>` + msg + `</p>` +
		`<p><a href="/">Voltar</a></p>` +
		`</main></body></html>`
}

func itoa(n int) string { // micro helper
	return strconv.Itoa(n)
}
