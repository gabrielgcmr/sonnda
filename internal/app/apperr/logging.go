// internal/app/apperr/logging.go
package apperr

import (
	"errors"
	"log/slog"
)

func ErrorCodeOf(err error) ErrorCode {
	var ae *AppError
	if errors.As(err, &ae) && ae != nil {
		return ae.Code
	}
	return INTERNAL_ERROR
}

func LogLevelOf(err error) slog.Level {
	code := ErrorCodeOf(err)

	switch code {
	// 5xx infra/internal -> Error
	case INFRA_AUTHENTICATION_ERROR,
		INFRA_DATABASE_ERROR,
		INFRA_STORAGE_ERROR,
		INFRA_EXTERNAL_SERVICE_ERROR,
		INFRA_TIMEOUT,
		INTERNAL_ERROR:
		return slog.LevelError

	// 429/413 -> Warn (abuso/limite)
	case RATE_LIMIT_EXCEEDED,
		UPLOAD_SIZE_EXCEEDED:
		return slog.LevelWarn

	// 4xx esperados -> Info
	default:
		return slog.LevelInfo
	}
}
