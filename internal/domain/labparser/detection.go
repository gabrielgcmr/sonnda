package labparser

import "strings"

func DetectExamType(rawText string) ExamType {
	text := NormalizeForMatch(rawText)
	if text == "" {
		return ExamTypeUnknown
	}

	if hasAnySignal(text, hemogramSignals) {
		return ExamTypeHemogram
	}

	return ExamTypeUnknown
}

func hasAnySignal(text string, signals []string) bool {
	for _, signal := range signals {
		if strings.Contains(text, signal) {
			return true
		}
	}
	return false
}

var hemogramSignals = []string{
	"hemograma",
	"eritrograma",
	"leucograma",
	"plaquetas",
}
