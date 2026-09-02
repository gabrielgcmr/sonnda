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

	lines := splitNonEmptyLines(input.RawText)
	usedLines := make(map[int]bool, len(lines))

	for index, line := range lines {
		result, ok := ParseHemogramLine(line)
		if ok {
			output.Results = append(output.Results, result)
			usedLines[index] = true
		}
	}

	output.Results = append(output.Results, parseHemogramBlocks(lines, usedLines)...)

	return output, nil
}

var (
	hemogramLinePattern       = regexp.MustCompile(`^\s*-?\s*(.+?)\s+([0-9OoIl.,]+)\s+(.+?)\s*$`)
	hemogramLineNoUnitPattern = regexp.MustCompile(`^\s*-?\s*(.+?)\s+([0-9OoIl.,]+)\s*$`)
)

type analyteLine struct {
	RawLine      string
	OriginalName string
	Code         string
}

type valueLine struct {
	RawLine  string
	RawValue string
	Value    float64
	Unit     string
}

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

func splitNonEmptyLines(text string) []string {
	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func parseHemogramBlocks(lines []string, usedLines map[int]bool) []ParsedLabResult {
	var results []ParsedLabResult

	for index := 0; index < len(lines); index++ {
		if usedLines[index] {
			continue
		}

		names, next := collectAnalyteBlock(lines, usedLines, index)
		if len(names) == 0 {
			continue
		}

		values, consumedUntil := collectValueBlock(lines, usedLines, next, len(names))
		if len(values) == 0 {
			index = next - 1
			continue
		}

		count := min(len(names), len(values))
		for offset := 0; offset < count; offset++ {
			results = append(results, resultFromBlockPair(names[offset], values[offset]))
		}

		index = consumedUntil - 1
	}

	return results
}

func collectAnalyteBlock(lines []string, usedLines map[int]bool, start int) ([]analyteLine, int) {
	var names []analyteLine

	for index := start; index < len(lines); index++ {
		if usedLines[index] {
			break
		}

		analyte, ok := parseAnalyteOnlyLine(lines[index])
		if !ok {
			break
		}
		names = append(names, analyte)
	}

	return names, start + len(names)
}

func collectValueBlock(lines []string, usedLines map[int]bool, start int, limit int) ([]valueLine, int) {
	var values []valueLine
	index := start

	for ; index < len(lines) && len(values) < limit; index++ {
		if usedLines[index] {
			break
		}
		if parseIgnoredBlockLine(lines[index]) {
			continue
		}

		value, ok := parseValueOnlyLine(lines[index])
		if !ok {
			if len(values) > 0 {
				break
			}
			continue
		}
		values = append(values, value)
	}

	return values, index
}

func parseAnalyteOnlyLine(rawLine string) (analyteLine, bool) {
	line := strings.TrimSpace(strings.TrimPrefix(NormalizeWhitespace(rawLine), "- "))
	if line == "" {
		return analyteLine{}, false
	}

	code, ok := ResolveAnalyteCode(line)
	if !ok {
		return analyteLine{}, false
	}

	return analyteLine{
		RawLine:      rawLine,
		OriginalName: line,
		Code:         code,
	}, true
}

func parseValueOnlyLine(rawLine string) (valueLine, bool) {
	body, _ := splitReference(NormalizeWhitespace(rawLine))
	matches := hemogramValueOnlyPattern.FindStringSubmatch(body)
	if len(matches) != 3 {
		return valueLine{}, false
	}

	rawValue := strings.TrimSpace(matches[1])
	value, err := ParseBrazilianDecimal(rawValue)
	if err != nil {
		return valueLine{}, false
	}

	return valueLine{
		RawLine:  rawLine,
		RawValue: rawValue,
		Value:    value,
		Unit:     strings.TrimSpace(matches[2]),
	}, true
}

var hemogramValueOnlyPattern = regexp.MustCompile(`^\s*([0-9OoIl.,]+)\s+(.+?)\s*$`)

func resultFromBlockPair(name analyteLine, value valueLine) ParsedLabResult {
	unit := value.Unit
	resultValue := value.Value

	return ParsedLabResult{
		Code:         name.Code,
		OriginalName: name.OriginalName,
		Value:        &resultValue,
		RawValue:     value.RawValue,
		Unit:         &unit,
		Status:       ParseStatusParsed,
		RawLine:      name.RawLine + "\n" + value.RawLine,
	}
}

func parseIgnoredBlockLine(line string) bool {
	normalized := NormalizeForMatch(line)
	switch normalized {
	case "",
		"valores de referencia",
		"valor de referencia",
		"percentual",
		"absoluta",
		"%":
		return true
	default:
		return false
	}
}
