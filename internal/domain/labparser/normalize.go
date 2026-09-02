package labparser

import (
	"fmt"
	"strconv"
	"strings"
)

func NormalizeWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func NormalizeForMatch(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c",
		"º", "", "ª", "", "³", "3",
		"-", " ", "_", " ", "/", " ", "\\", " ",
		"(", " ", ")", " ", "[", " ", "]", " ",
		":", " ", ";", " ", ",", " ",
		"\n", " ", "\r", " ", "\t", " ",
	)
	return NormalizeWhitespace(replacer.Replace(text))
}

func ParseBrazilianDecimal(raw string) (float64, error) {
	normalized := NormalizeNumericToken(raw)
	if normalized == "" {
		return 0, fmt.Errorf("empty numeric token")
	}

	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric token %q: %w", raw, err)
	}
	return value, nil
}

func NormalizeNumericToken(raw string) string {
	text := strings.TrimSpace(raw)
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "\u00a0", "")
	if text == "" {
		return ""
	}

	// Correcoes restritas a tokens que ja foram identificados como numericos.
	text = strings.NewReplacer(
		"O", "0",
		"o", "0",
		"I", "1",
		"l", "1",
	).Replace(text)

	if strings.Contains(text, ",") {
		text = strings.ReplaceAll(text, ".", "")
		text = strings.ReplaceAll(text, ",", ".")
		return text
	}

	if looksLikeThousandsWithDots(text) {
		return strings.ReplaceAll(text, ".", "")
	}

	return text
}

func looksLikeThousandsWithDots(text string) bool {
	if !strings.Contains(text, ".") {
		return false
	}

	parts := strings.Split(text, ".")
	if len(parts) < 2 || parts[0] == "" {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}

	for _, part := range parts[1:] {
		if len(part) != 3 {
			return false
		}
	}
	return true
}
