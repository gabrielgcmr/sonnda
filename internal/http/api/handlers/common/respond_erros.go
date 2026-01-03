// internal/http/api/handlers/common/errors.go
package common

import (
	"errors"
	"net/http"

	applog "sonnda-api/internal/app/observability"
	"sonnda-api/internal/domain/model/medicalrecord/labs"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/shared"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"
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

	switch {
	// auth
	case errors.Is(err, user.ErrInvalidAuthProvider),
		errors.Is(err, user.ErrInvalidAuthSubject),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrInvalidRole),
		errors.Is(err, user.ErrInvalidFullName),
		errors.Is(err, user.ErrInvalidBirthDate),
		errors.Is(err, user.ErrInvalidCPF),
		errors.Is(err, user.ErrInvalidPhone):
		RespondError(c, http.StatusBadRequest, "invalid_auth", err)
	case errors.Is(err, user.ErrEmailAlreadyExists):
		RespondError(c, http.StatusConflict, "email_already_exists", err)
	case errors.Is(err, user.ErrCPFAlreadyExists):
		RespondError(c, http.StatusConflict, "cpf_already_exists", err)
	case errors.Is(err, user.ErrAuthIdentityAlreadyExists):
		RespondError(c, http.StatusConflict, "auth_identity_already_exists", err)
	case errors.Is(err, user.ErrIdentityAlreadyLinkedError):
		RespondError(c, http.StatusConflict, "identity_already_linked", err)

	// professional profile
	case errors.Is(err, professional.ErrRegistrationRequired),
		errors.Is(err, professional.ErrInvalidUserID),
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer):
		RespondError(c, http.StatusBadRequest, "invalid_professional_registration", err)
	case errors.Is(err, professional.ErrProfileNotFound):
		RespondError(c, http.StatusNotFound, "professional_profile_not_found", err)

	// authorization
	case errors.Is(err, user.ErrAuthorizationForbidden):
		RespondError(c, http.StatusForbidden, "forbidden", nil)
	// patient
	case errors.Is(err, patient.ErrPatientNotFound):
		RespondError(c, http.StatusNotFound, "patient_not_found", nil)
	case errors.Is(err, patient.ErrCPFAlreadyExists):
		RespondError(c, http.StatusConflict, "cpf_already_exists", nil)
	case errors.Is(err, shared.ErrInvalidBirthDate):
		RespondError(c, http.StatusBadRequest, "invalid_birth_date", err)
	case errors.Is(err, shared.ErrInvalidGender):
		RespondError(c, http.StatusBadRequest, "invalid_gender", err)
	case errors.Is(err, shared.ErrInvalidRace):
		RespondError(c, http.StatusBadRequest, "invalid_race", err)

	// labs
	case errors.Is(err, labs.ErrLabReportNotFound):
		RespondError(c, http.StatusNotFound, "lab_report_not_found", nil)
	case errors.Is(err, labs.ErrLabReportAlreadyExists):
		RespondError(c, http.StatusConflict, "lab_report_already_exists", nil)
	case errors.Is(err, labs.ErrInvalidDocument):
		RespondError(c, http.StatusBadRequest, "invalid_document", err)
	case errors.Is(err, labs.ErrInvalidInput), errors.Is(err, labs.ErrMissingId):
		RespondError(c, http.StatusBadRequest, "invalid_input", err)
	case errors.Is(err, labs.ErrDocumentProcessing):
		// serviço externo falhou -> 502
		RespondError(c, http.StatusBadGateway, "document_processing_failed", err)
	case errors.Is(err, labs.ErrInvalidDateFormat):
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
