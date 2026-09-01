package examsvc

import (
	"regexp"
	"strings"
)

type reportMetadata struct {
	Title      *string
	Modality   *string
	BodySite   *string
	Conclusion *string
}

func inferReportMetadata(reportText string) reportMetadata {
	title := inferReportTitle(reportText)
	return reportMetadata{
		Title:      title,
		Modality:   inferModality(title, reportText),
		BodySite:   inferBodySite(title, reportText),
		Conclusion: extractConclusion(reportText),
	}
}

func inferReportTitle(reportText string) *string {
	for _, line := range reportLines(reportText) {
		normalized := normalizeRouteText(line)
		if looksLikeExamTitle(normalized) {
			return &line
		}
	}
	return nil
}

func looksLikeExamTitle(normalizedLine string) bool {
	titleSignals := []string{
		"ultrassonografia",
		"ultrassom",
		"tomografia",
		"ressonancia",
		"radiografia",
		"raio x",
		"mamografia",
		"hemograma",
		"exame",
	}

	for _, signal := range titleSignals {
		if strings.Contains(normalizedLine, signal) {
			return true
		}
	}
	return false
}

func inferModality(title *string, reportText string) *string {
	text := normalizeRouteText(joinOptional(title, reportText))

	rules := []struct {
		signal string
		value  string
	}{
		{signal: "ultrassonografia", value: "ultrasound"},
		{signal: "ultrassom", value: "ultrasound"},
		{signal: "tomografia", value: "ct"},
		{signal: "ressonancia", value: "mri"},
		{signal: "radiografia", value: "xray"},
		{signal: "raio x", value: "xray"},
		{signal: "mamografia", value: "mammography"},
		{signal: "hemograma", value: "blood_test"},
	}

	for _, rule := range rules {
		if strings.Contains(text, rule.signal) {
			value := rule.value
			return &value
		}
	}
	return nil
}

func inferBodySite(title *string, reportText string) *string {
	text := normalizeRouteText(joinOptional(title, reportText))

	rules := []struct {
		signal string
		value  string
	}{
		{signal: "aparelho urinario", value: "aparelho_urinario"},
		{signal: "urinario", value: "aparelho_urinario"},
		{signal: "prostatica", value: "prostata"},
		{signal: "prostata", value: "prostata"},
		{signal: "abdome", value: "abdome"},
		{signal: "abdominal", value: "abdome"},
		{signal: "torax", value: "torax"},
		{signal: "cranio", value: "cranio"},
		{signal: "joelho", value: "joelho"},
		{signal: "mama", value: "mamas"},
		{signal: "tireoide", value: "tireoide"},
	}

	for _, rule := range rules {
		if strings.Contains(text, rule.signal) {
			value := rule.value
			return &value
		}
	}
	return nil
}

func extractConclusion(reportText string) *string {
	lines := reportLines(reportText)
	for index, line := range lines {
		normalized := normalizeRouteText(strings.TrimSuffix(line, ":"))
		if normalized != "opiniao" && normalized != "conclusao" && normalized != "impressao diagnostica" {
			continue
		}

		// Guarda o bloco textual mais util para leitura do paciente.
		block := collectSectionBlock(lines, index+1)
		if block == "" {
			continue
		}
		return &block
	}
	return nil
}

func collectSectionBlock(lines []string, start int) string {
	var block []string
	for _, line := range lines[start:] {
		normalized := normalizeRouteText(strings.TrimSuffix(line, ":"))
		if isSectionHeader(normalized) && len(block) > 0 {
			break
		}
		block = append(block, line)
	}

	text := strings.TrimSpace(strings.Join(block, "\n"))
	if text == "" {
		return ""
	}
	return text
}

func isSectionHeader(normalizedLine string) bool {
	if normalizedLine == "" {
		return false
	}
	headers := []string{
		"metodologia",
		"analise",
		"medidas",
		"opiniao",
		"conclusao",
		"impressao diagnostica",
	}
	for _, header := range headers {
		if normalizedLine == header {
			return true
		}
	}
	return false
}

func reportLines(text string) []string {
	rawLines := regexp.MustCompile(`\r?\n`).Split(text, -1)
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func joinOptional(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value + " " + fallback
}
