package labparser

import (
	"regexp"
	"strings"
)

var referenceNumberPattern = regexp.MustCompile(`[0-9OoIl.,]+`)

func ParseReferenceRange(raw string) ReferenceRange {
	text := NormalizeWhitespace(strings.TrimSpace(raw))
	if text == "" {
		return ReferenceRange{}
	}

	matches := referenceNumberPattern.FindAllString(normalizeReferenceForNumbers(text), -1)
	if len(matches) != 2 {
		return ReferenceRange{Text: text}
	}

	min, minErr := ParseBrazilianDecimal(matches[0])
	max, maxErr := ParseBrazilianDecimal(matches[1])
	if minErr != nil || maxErr != nil {
		return ReferenceRange{Text: text}
	}

	return ReferenceRange{
		Min:  &min,
		Max:  &max,
		Text: text,
	}
}

func normalizeReferenceForNumbers(text string) string {
	return strings.NewReplacer(
		"mm³", "mm",
		"mm3", "mm",
		"MM³", "MM",
		"MM3", "MM",
	).Replace(text)
}
