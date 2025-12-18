package handlers

import (
	"log/slog"
	"net/http"
	"sonnda-api/internal/core/domain"
	"strings"

	"github.com/gin-gonic/gin"
)

// RespondError escreve uma resposta JSON padrão e registra log estruturado.
// Refinamento de nível:
// - 5xx -> Error
// - 401/403 -> Info
// - outros 4xx -> Warn (ou Info para invalid_input, se você quiser)
func RespondError(c *gin.Context, log *slog.Logger, status int, code string, err error) {
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

func RespondDomainError(c *gin.Context, log *slog.Logger, err error) {
	switch err {
	// auth
	case domain.ErrInvalidAuthProvider, domain.ErrInvalidAuthSubject, domain.ErrInvalidEmail, domain.ErrInvalidRole:
		RespondError(c, log, http.StatusBadRequest, "invalid_auth", err)
	case domain.ErrEmailAlreadyExists:
		RespondError(c, log, http.StatusConflict, "email_already_exists", err)

	// authorization
	case domain.ErrForbidden:
		RespondError(c, log, http.StatusForbidden, "forbidden", nil)

	// patient
	case domain.ErrPatientNotFound:
		RespondError(c, log, http.StatusNotFound, "patient_not_found", nil)
	case domain.ErrCPFAlreadyExists:
		RespondError(c, log, http.StatusConflict, "cpf_already_exists", nil)
	case domain.ErrInvalidBirthDate:
		RespondError(c, log, http.StatusBadRequest, "invalid_birth_date", err)

	// labs
	case domain.ErrLabReportNotFound:
		RespondError(c, log, http.StatusNotFound, "lab_report_not_found", nil)
	case domain.ErrLabReportAlreadyExists:
		RespondError(c, log, http.StatusConflict, "lab_report_already_exists", nil)
	case domain.ErrInvalidDocument:
		RespondError(c, log, http.StatusBadRequest, "invalid_document", err)
	case domain.ErrInvalidInput, domain.ErrMissingIdentifiers:
		RespondError(c, log, http.StatusBadRequest, "invalid_input", err)
	case domain.ErrDocumentProcessing:
		// serviço externo falhou -> 502
		RespondError(c, log, http.StatusBadGateway, "document_processing_failed", err)
	case domain.ErrInvalidDateFormat:
		RespondError(c, log, http.StatusBadRequest, "invalid_date_format", err)

	default:
		// fallback: não vazar detalhes em 5xx (seu RespondError já protege isso)
		RespondError(c, log, http.StatusInternalServerError, "server_error", err)
	}
}

func RespondUploadError(c *gin.Context, log *slog.Logger, err error) {
	msg := err.Error()

	switch {
	case strings.HasPrefix(msg, "file_required"):
		// inclui o caso: "request Content-Type isn't multipart/form-data"
		RespondError(c, log, http.StatusBadRequest, "file_required", err)

	case msg == "empty_file":
		RespondError(c, log, http.StatusBadRequest, "empty_file", nil)

	case msg == "file_too_large":
		RespondError(c, log, http.StatusRequestEntityTooLarge, "file_too_large", nil)

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
		RespondError(c, log, http.StatusInternalServerError, "open_file_failed", err)

	case strings.HasPrefix(msg, "upload_failed"):
		RespondError(c, log, http.StatusBadGateway, "upload_failed", err)

	default:
		// fallback
		RespondError(c, log, http.StatusBadRequest, "upload_failed", err)
	}
}
