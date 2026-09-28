// cmd/extract-lab-summary/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	geminiinfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/gemini"
	"github.com/joho/godotenv"
)

func main() {
	inputPath := flag.String("input", "", "OCR text file to summarize")
	outputPath := flag.String("output", "", "summary output file; defaults to stdout")
	flag.Parse()

	if strings.TrimSpace(*inputPath) == "" {
		exitWithError(2, errors.New("missing required -input"))
	}
	if err := summarizeFile(*inputPath, *outputPath); err != nil {
		exitWithError(1, err)
	}
}

func summarizeFile(inputPath, outputPath string) error {
	text, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read OCR text: %w", err)
	}
	if strings.TrimSpace(string(text)) == "" {
		return errors.New("OCR text file is empty")
	}

	_ = godotenv.Load()
	geminiConfig, err := config.LoadGemini()
	if err != nil {
		return fmt.Errorf("load Gemini configuration: %w", err)
	}

	ctx := context.Background()
	client, err := geminiinfra.NewClient(ctx, geminiConfig)
	if err != nil {
		return fmt.Errorf("create Gemini client: %w", err)
	}
	extractor, err := geminiinfra.NewLabReportTextExtractor(client)
	if err != nil {
		return fmt.Errorf("create lab report extractor: %w", err)
	}
	report, err := extractor.ExtractLabReport(ctx, labextraction.ExtractLabReportInput{Text: string(text)})
	if err != nil {
		return fmt.Errorf("extract lab results: %w", err)
	}

	summary := formatLabSummary(report)
	if strings.TrimSpace(outputPath) == "" {
		_, err = fmt.Print(summary)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(summary), 0o644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	fmt.Fprintf(os.Stderr, "ok %s -> %s\n", inputPath, outputPath)
	return nil
}

func exitWithError(code int, err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}
