package labparser

import (
	"context"
	"regexp"
	"strings"
)

type HemogramParser struct{}

func NewHemogramParser() *HemogramParser {
	return &HemogramParser{}
}

func (p *HemogramParser) Parse(ctx context.Context, input ParseInput) (*ParseOutput, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	output := &ParseOutput{
		RawText:        input.RawText,
		NormalizedText: NormalizeForMatch(input.RawText),
		ExamType:       ExamTypeHemogram,
	}

	for _, line := range strings.Split(input.RawText, "\n") {
		result, ok := ParseHemogramLine(line)
		if ok {
			output.Results = append(output.Results, result)
		}
	}

	return output, nil
}

var (
	hemogramLinePattern       = regexp.MustCompile(`^\s*-?\s*(.+?)\s+([0-9OoIl.,]+)\s+(.+?)\s*$`)
	hemogramLineNoUnitPattern = regexp.MustCompile(`^\s*-?\s*(.+?)\s+([0-9OoIl.,]+)\s*$`)
)

func ParseHemogramLine(rawLine string) (ParsedLabResult, bool) {
	line := NormalizeWhitespace(rawLine)
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimSpace(line)
	if line == "" {
		return ParsedLabResult{}, false
	}

	body, rawReference := splitReference(line)
	matches := hemogramLinePattern.FindStringSubmatch(body)
	if len(matches) != 4 {
		matches = hemogramLineNoUnitPattern.FindStringSubmatch(body)
		if len(matches) != 3 {
			return ParsedLabResult{}, false
		}
	}

	originalName := strings.TrimSpace(matches[1])
	rawValue := strings.TrimSpace(matches[2])
	var unit string
	if len(matches) == 4 {
		unit = strings.TrimSpace(matches[3])
	}
	value, valueErr := ParseBrazilianDecimal(rawValue)
	code, codeOK := ResolveAnalyteCode(originalName)
	if !codeOK && unit == "" {
		return ParsedLabResult{}, false
	}

	result := ParsedLabResult{
		Code:         code,
		OriginalName: originalName,
		RawValue:     rawValue,
		RawLine:      rawLine,
		Status:       ParseStatusParsed,
	}
	if unit != "" {
		result.Unit = &unit
	}
	if valueErr == nil {
		result.Value = &value
	}
	if rawReference != "" {
		reference := ParseReferenceRange(rawReference)
		result.RawReference = &reference.Text
		result.ReferenceMin = reference.Min
		result.ReferenceMax = reference.Max
	}

	if !codeOK || valueErr != nil || unit == "" {
		result.Status = ParseStatusPartial
	}

	return result, true
}

func splitReference(line string) (body string, rawReference string) {
	start := strings.Index(line, "(")
	end := strings.LastIndex(line, ")")
	if start < 0 {
		return strings.TrimSpace(line), ""
	}

	body = strings.TrimSpace(line[:start])
	reference := strings.TrimSpace(line[start+1:])
	if end > start {
		reference = strings.TrimSpace(line[start+1 : end])
	}
	normalized := NormalizeForMatch(reference)
	switch {
	case strings.HasPrefix(normalized, "referencia "):
		reference = strings.TrimSpace(reference[len(referenceLabel(reference)):])
	case normalized == "referencia":
		reference = ""
	}

	return body, strings.TrimSpace(reference)
}

func referenceLabel(reference string) string {
	if idx := strings.Index(reference, ":"); idx >= 0 {
		return reference[:idx+1]
	}
	return "Referencia"
}
