package labs

import (
	"strings"
	"time"

	"sonnda-api/internal/core/domain"
)

// Helpers de parsing de datas
func (uc *ExtractLabsUseCase) parseDate(s string) (time.Time, error) {
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

func (uc *ExtractLabsUseCase) parseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// normalizacoes simples que costumam vir em laudos ("as", "h")
	s = strings.ReplaceAll(s, "\u00e0s", " ")
	s = strings.ReplaceAll(s, " as ", " ")
	s = strings.ReplaceAll(s, "h", ":")
	s = strings.Join(strings.Fields(s), " ")

	layouts := []string{
		time.RFC3339,                // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,            // "2006-01-02T15:04:05.999Z07:00"
		"2006-01-02 15:04:05",       // "2025-12-03 14:30:00"
		"2006-01-02 15:04",          // "2025-12-03 14:30"
		"2006-01-02 15:04:05-07:00", // "2025-12-03 14:30:00-03:00"
		"02/01/2006 15:04:05",       // "03/12/2025 14:30:00"
		"02/01/2006 15:04",          // "03/12/2025 14:30"
		"02/01/2006 15:04:05 -0700", // "03/12/2025 14:30:00 -0300"
		"02/01/2006 15:04 -0700",    // "03/12/2025 14:30 -0300"
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	// Tenta como data simples
	return uc.parseDate(s)
}

func normalize(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	// aqui dá pra remover acentos se quiser ser mais agressivo
	// ex: usar norm.NFD + remover runas com categoria Mn
	return s
}

func normalizeValue(v *string) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(*v)
	// se quiser, limpa espaços, troca vírgula por ponto, etc.
	return s
}
