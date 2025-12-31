// internal/http/api/handlers/common/errors.go
package common

import (
	"net/http"

	applog "sonnda-api/internal/app/observability"
	"sonnda-api/internal/domain/entities/medicalrecord/lab"
	"sonnda-api/internal/domain/entities/patient"
	"sonnda-api/internal/domain/entities/professional"
	"sonnda-api/internal/domain/entities/shared"
	"sonnda-api/internal/domain/entities/user"
	"strings"

	"github.com/gin-gonic/gin"
)

// RespondError escreve uma resposta JSON padrão e registra log estruturado.
// Refinamento de nível:
// - 5xx -> Error
// - 401/403 -> Info
// - outros 4xx -> Warn (ou Info para invalid_input, se você quiser)
func RespondError(c *gin.Context, status int, code string, err error) {
	//Registra o erro internament no Gim
	if err != nil {
		_ = c.Error(err)
	}
	// Salva o código de erro no contexto (útil para middlewares de logging)
	c.Set("error_code", code)

	log := applog.FromContext(c.Request.Context())
	// Decide nível
	switch {
	case status >= 500:
		if err != nil {
			log.Error("handler_error", "status", status, "error", code, "err", err)
		} else {
			log.Error("handler_error", "status", status, "error", code)
		}
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		if err != nil {
			log.Info("handler_error", "status", status, "error", code, "err", err)
		} else {
			log.Info("handler_error", "status", status, "error", code)
		}
	default:
		// Opcional: invalid_input como Info para reduzir ruído
		if status == http.StatusBadRequest && code == "invalid_input" {
			if err != nil {
				log.Info("handler_error", "status", status, "error", code, "err", err)
			} else {
				log.Info("handler_error", "status", status, "error", code)
			}
			break
		}

		if err != nil {
			log.Warn("handler_error", "status", status, "error", code, "err", err)
		} else {
			log.Warn("handler_error", "status", status, "error", code)
		}
	}

	// Resposta
	// ⚠️ Não vazar detalhes internos em 5xx: loga o err, mas responde genérico.
	if status >= 500 {
		c.JSON(status, gin.H{"error": code})
		return
	}

	if err != nil {
		c.JSON(status, gin.H{
			"error":   code,
			"details": err.Error(),
		})
		return
	}

	c.JSON(status, gin.H{"error": code})
}

func RespondDomainError(c *gin.Context, err error) {

	switch err {
	// auth
	case user.ErrInvalidAuthProvider,
		user.ErrInvalidAuthSubject,
		user.ErrInvalidEmail,
		user.ErrInvalidRole,
		user.ErrInvalidFullName,
		user.ErrInvalidBirthDate,
		user.ErrInvalidCPF,
		user.ErrInvalidPhone:
		RespondError(c, http.StatusBadRequest, "invalid_auth", err)
	case user.ErrEmailAlreadyExists:
		RespondError(c, http.StatusConflict, "email_already_exists", err)
	case user.ErrAuthIdentityAlreadyExists:
		RespondError(c, http.StatusConflict, "auth_identity_already_exists", err)
	case user.ErrIdentityAlreadyLinkedError:
		RespondError(c, http.StatusConflict, "identity_already_linked", err)

	// professional profile
	case professional.ErrRegistrationRequired,
		professional.ErrInvalidUserID,
		professional.ErrInvalidRegistrationNumber,
		professional.ErrInvalidRegistrationIssuer:
		RespondError(c, http.StatusBadRequest, "invalid_professional_registration", err)
	case professional.ErrProfileNotFound:
		RespondError(c, http.StatusNotFound, "professional_profile_not_found", err)

	// authorization
	case user.ErrAuthorizationForbidden:
		RespondError(c, http.StatusForbidden, "forbidden", nil)
	// patient
	case patient.ErrPatientNotFound:
		RespondError(c, http.StatusNotFound, "patient_not_found", nil)
	case patient.ErrCPFAlreadyExists:
		RespondError(c, http.StatusConflict, "cpf_already_exists", nil)
	case shared.ErrInvalidBirthDate:
		RespondError(c, http.StatusBadRequest, "invalid_birth_date", err)

	// labs
	case lab.ErrLabReportNotFound:
		RespondError(c, http.StatusNotFound, "lab_report_not_found", nil)
	case lab.ErrLabReportAlreadyExists:
		RespondError(c, http.StatusConflict, "lab_report_already_exists", nil)
	case lab.ErrInvalidDocument:
		RespondError(c, http.StatusBadRequest, "invalid_document", err)
	case lab.ErrInvalidInput, lab.ErrMissingId:
		RespondError(c, http.StatusBadRequest, "invalid_input", err)
	case lab.ErrDocumentProcessing:
		// serviço externo falhou -> 502
		RespondError(c, http.StatusBadGateway, "document_processing_failed", err)
	case lab.ErrInvalidDateFormat:
		RespondError(c, http.StatusBadRequest, "invalid_date_format", err)

	default:
		// fallback: não vazar detalhes em 5xx (seu RespondError já protege isso)
		RespondError(c, http.StatusInternalServerError, "server_error", err)
	}
}

func RespondUploadError(c *gin.Context, err error) {

	log := applog.FromContext(c.Request.Context())

	msg := err.Error()

	switch {
	case strings.HasPrefix(msg, "file_required"):
		// inclui o caso: "request Content-Type isn't multipart/form-data"
		RespondError(c, http.StatusBadRequest, "file_required", err)

	case msg == "empty_file":
		RespondError(c, http.StatusBadRequest, "empty_file", nil)

	case msg == "file_too_large":
		RespondError(c, http.StatusRequestEntityTooLarge, "file_too_large", nil)

	case strings.HasPrefix(msg, "unsupported_mime_type:"):
		ct := strings.TrimPrefix(msg, "unsupported_mime_type:")
		// aqui eu devolvo payload mais útil pro front
		log.Info("unsupported_mime_type", "content_type", ct)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "unsupported_mime_type",
			"content_type":  ct,
			"allowed_types": []string{"application/pdf", "image/jpeg", "image/png"},
		})

	case strings.HasPrefix(msg, "open_file_failed"):
		RespondError(c, http.StatusInternalServerError, "open_file_failed", err)

	case strings.HasPrefix(msg, "upload_failed"):
		RespondError(c, http.StatusBadGateway, "upload_failed", err)

	default:
		// fallback
		RespondError(c, http.StatusBadRequest, "upload_failed", err)
	}
}
