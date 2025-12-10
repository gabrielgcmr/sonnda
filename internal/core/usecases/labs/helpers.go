package labs

import (
	"sonnda-api/internal/core/domain"
	"strings"
	"time"
)

// Helpers de parsing de datas
func (uc *CreateFromDocumentUseCase) parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	layouts := []string{
		"2006-01-02", // ISO: "2025-12-03"
		"02/01/2006", // BR: "03/12/2025"
		"2006/01/02", // Alt: "2025/12/03"
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, domain.ErrInvalidDateFormat
}

func (uc *CreateFromDocumentUseCase) parseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	layouts := []string{
		time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,      // "2006-01-02T15:04:05.999Z07:00"
		"2006-01-02 15:04:05", // "2025-12-03 14:30:00"
		"2006-01-02 15:04",    // "2025-12-03 14:30"
		"02/01/2006 15:04:05", // "03/12/2025 14:30:00"
		"02/01/2006 15:04",    // "03/12/2025 14:30"
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	// Tenta como data simples
	return uc.parseDate(s)
}
