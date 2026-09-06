// cmd/lab-extraction-eval/main.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/gemini"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/joho/godotenv"
)

func main() {
	var inputPath string
	var expectedPath string
	var forceLab bool
	flag.StringVar(&inputPath, "input", "", "path to extracted lab text fixture")
	flag.StringVar(&expectedPath, "expected", "", "optional path to expected JSON")
	flag.BoolVar(&forceLab, "force-lab", false, "call Gemini even when the local route is not laboratory")
	flag.Parse()

	if inputPath == "" {
		exitWithError(2, errors.New("missing required -input"))
	}

	text, err := os.ReadFile(inputPath)
	if err != nil {
		exitWithError(1, fmt.Errorf("read input: %w", err))
	}

	route := examsvc.NewHeuristicExamRouter().Route(examsvc.ExamRouteInput{
		ExtractedText:    string(text),
		MimeType:         "text/plain",
		OriginalFilename: filepath.Base(inputPath),
	})
	if route.ExamType != exams.ExamTypeLaboratory && !forceLab {
		writeSkip(route)
		return
	}

	_ = godotenv.Load()
	cfg, err := config.LoadGemini()
	if err != nil {
		exitWithConfigError(err)
	}

	ctx := context.Background()
	client, err := gemini.NewClient(ctx, cfg)
	if err != nil {
		exitWithError(1, fmt.Errorf("create gemini client: %w", err))
	}
	extractor, err := gemini.NewLabReportTextExtractor(client)
	if err != nil {
		exitWithError(1, fmt.Errorf("create lab extractor: %w", err))
	}

	report, err := extractor.ExtractLabReport(ctx, labextraction.ExtractLabReportInput{Text: string(text)})
	if err != nil {
		exitWithError(1, fmt.Errorf("extract lab report: %w", err))
	}

	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		exitWithError(1, fmt.Errorf("encode extracted report: %w", err))
	}
	fmt.Println(string(output))

	if expectedPath == "" {
		return
	}
	matched, err := compareJSON(output, expectedPath)
	if err != nil {
		exitWithError(1, err)
	}
	if !matched {
		fmt.Fprintln(os.Stderr, "eval mismatch: generated JSON differs from expected JSON")
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "eval ok: generated JSON matches expected JSON")
}

func writeSkip(route examsvc.ExamRouteResult) {
	output := struct {
		Skipped        bool                 `json:"skipped"`
		Reason         string               `json:"reason"`
		ExamType       exams.ExamType       `json:"exam_type"`
		Status         exams.DocumentStatus `json:"status"`
		Confidence     float64              `json:"confidence"`
		MatchedSignals []string             `json:"matched_signals,omitempty"`
	}{
		Skipped:        true,
		Reason:         "local_gate_not_laboratory",
		ExamType:       route.ExamType,
		Status:         route.Status,
		Confidence:     route.Confidence,
		MatchedSignals: route.MatchedSignals,
	}
	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		exitWithError(1, fmt.Errorf("encode skip output: %w", err))
	}
	fmt.Println(string(encoded))
	fmt.Fprintln(os.Stderr, "Gemini call skipped; use -force-lab to bypass the local gate.")
}

func compareJSON(generated []byte, expectedPath string) (bool, error) {
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		return false, fmt.Errorf("read expected: %w", err)
	}
	var generatedDoc any
	if err := json.Unmarshal(generated, &generatedDoc); err != nil {
		return false, fmt.Errorf("decode generated JSON: %w", err)
	}
	var expectedDoc any
	if err := json.Unmarshal(expected, &expectedDoc); err != nil {
		return false, fmt.Errorf("decode expected JSON: %w", err)
	}
	return reflect.DeepEqual(generatedDoc, expectedDoc), nil
}

func exitWithConfigError(err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil && len(appErr.Violations) > 0 {
		fmt.Fprintln(os.Stderr, appErr.Message)
		for _, violation := range appErr.Violations {
			fmt.Fprintf(os.Stderr, " - %s: %s\n", violation.Field, violation.Reason)
		}
		os.Exit(1)
	}
	exitWithError(1, err)
}

func exitWithError(code int, err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}
