// cmd/lab-extraction-eval/main.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/gemini"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/joho/godotenv"
)

func main() {
	var inputPath string
	var expectedPath string
	flag.StringVar(&inputPath, "input", "", "path to extracted lab text fixture")
	flag.StringVar(&expectedPath, "expected", "", "optional path to expected JSON")
	flag.Parse()

	if inputPath == "" {
		exitWithError(2, errors.New("missing required -input"))
	}

	_ = godotenv.Load()
	cfg, err := config.LoadGemini()
	if err != nil {
		exitWithConfigError(err)
	}

	text, err := os.ReadFile(inputPath)
	if err != nil {
		exitWithError(1, fmt.Errorf("read input: %w", err))
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
