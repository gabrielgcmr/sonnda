// internal/config/ocr.go
package config

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type OCRConfig struct {
	Timeout time.Duration
}

func loadOCRConfig() (OCRConfig, error) {
	timeout, err := time.ParseDuration(getEnvOrDefault("OCR_TIMEOUT", "60s"))
	if err != nil || timeout <= 0 {
		return OCRConfig{}, apperr.Validation("invalid configuration", apperr.Violation{Field: "OCR_TIMEOUT", Reason: "must_be_positive_duration"})
	}
	return OCRConfig{Timeout: timeout}, nil
}
