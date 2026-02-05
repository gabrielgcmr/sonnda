// internal/api/problem/problem.go
package problem

import (
	"net/http"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

// Details implementa RFC 9457 (Problem Details for HTTP APIs).
// "code", "violations" e "request_id" são extension members.
type Details struct {
	Type       string             `json:"type"`
	Title      string             `json:"title"`
	Status     int                `json:"status"`
	Detail     string             `json:"detail"`
	Instance   string             `json:"instance,omitempty"`
	Code       apperr.ErrorCode   `json:"code"`
	Violations []apperr.Violation `json:"violations,omitempty"`
	RequestID  string             `json:"request_id,omitempty"`
}

type Meta struct {
	Instance  string
	RequestID string
}

func New(status int, code apperr.ErrorCode, detail string, violations []apperr.Violation, meta Meta) Details {
	return Details{
		Type:       TypeFromCode(code),
		Title:      TitleFromCode(code, status),
		Status:     status,
		Detail:     detail,
		Instance:   strings.TrimSpace(meta.Instance),
		Code:       code,
		Violations: violations,
		RequestID:  strings.TrimSpace(meta.RequestID),
	}
}

func TypeFromCode(code apperr.ErrorCode) string {
	c := strings.TrimSpace(string(code))
	if c == "" {
		return "about:blank"
	}
	// URN estável e válido como URI.
	return "urn:sonnda:problem:" + strings.ToLower(c)
}

func TitleFromCode(code apperr.ErrorCode, status int) string {
	switch code {
	// AUTH
	case apperr.AUTH_REQUIRED, apperr.AUTH_TOKEN_INVALID, apperr.AUTH_TOKEN_EXPIRED:
		return "Não autorizado"

	// AUTHZ
	case apperr.ACCESS_DENIED, apperr.ACTION_NOT_ALLOWED:
		return "Acesso negado"

	// VALIDATION
	case apperr.VALIDATION_FAILED,
		apperr.REQUIRED_FIELD_MISSING,
		apperr.INVALID_FIELD_FORMAT,
		apperr.INVALID_ENUM_VALUE,
		apperr.INVALID_DATE:
		return "Falha de validação"

	// NOT FOUND
	case apperr.NOT_FOUND:
		return "Não encontrado"

	// CONFLICT
	case apperr.RESOURCE_CONFLICT, apperr.RESOURCE_ALREADY_EXISTS:
		return "Conflito"

	// DOMAIN
	case apperr.DOMAIN_RULE_VIOLATION:
		return "Regra de negócio violada"

	// RATE
	case apperr.RATE_LIMIT_EXCEEDED:
		return "Muitas requisições"
	case apperr.UPLOAD_SIZE_EXCEEDED:
		return "Payload muito grande"

	// INFRA / INTERNAL
	case apperr.INFRA_EXTERNAL_SERVICE_ERROR,
		apperr.INFRA_TIMEOUT,
		apperr.INFRA_AUTHENTICATION_ERROR,
		apperr.INFRA_DATABASE_ERROR,
		apperr.INFRA_STORAGE_ERROR,
		apperr.INTERNAL_ERROR:
		return "Erro interno"
	}

	if t := strings.TrimSpace(http.StatusText(status)); t != "" {
		return t
	}
	return "Erro"
}
