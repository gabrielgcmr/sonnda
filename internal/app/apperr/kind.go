// internal/app/apperr/kind.go
package apperr

// Kind representa a categoria semântica do erro.
// NÃO representa diretamente HTTP status.
type Kind string

const (
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindInvalidInput Kind = "invalid_input"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"

	// Infra / dependências externas
	KindExternal    Kind = "external"
	KindBadGateway  Kind = "bad_gateway"
	KindUnavailable Kind = "unavailable"
	KindTimeout     Kind = "timeout"

	// Operacional / fallback
	KindServiceClosed Kind = "service_closed"
	KindInternal      Kind = "internal"
)
