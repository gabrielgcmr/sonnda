// cmd/extract-text/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	textextractioninfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/textextraction"
)

func main() {
	var inputPath string
	var outputDir string
	var requireUsable bool
	flag.StringVar(&inputPath, "input", "", "file or directory with PDFs/images")
	flag.StringVar(&outputDir, "output", "", "directory for generated .txt files")
	flag.BoolVar(&requireUsable, "require-usable", false, "fail when extracted text does not pass the runtime quality filter")
	flag.Parse()

	if strings.TrimSpace(inputPath) == "" {
		exitWithError(2, errors.New("missing required -input"))
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		exitWithError(1, fmt.Errorf("stat input: %w", err))
	}

	extractor := textextractioninfra.NewCommandExtractorWithOptions(textextractioninfra.CommandExtractorOptions{
		RequireUsableText: requireUsable,
	})
	ctx := context.Background()

	if !info.IsDir() {
		if err := extractOne(ctx, extractor, inputPath, inputPath, outputDir); err != nil {
			exitWithError(1, err)
		}
		return
	}

	if outputDir == "" {
		outputDir = filepath.Join("samples", "text")
	}

	var failed int
	err = filepath.WalkDir(inputPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			failed++
			fmt.Fprintf(os.Stderr, "fail %s: %v\n", path, walkErr)
			return nil
		}
		if entry.IsDir() || mimeTypeFromPath(path) == "" {
			return nil
		}
		if err := extractOne(ctx, extractor, inputPath, path, outputDir); err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "fail %s: %v\n", path, err)
		}
		return nil
	})
	if err != nil {
		exitWithError(1, fmt.Errorf("walk input: %w", err))
	}
	if failed > 0 {
		exitWithError(1, fmt.Errorf("finished with %d extraction failure(s)", failed))
	}
}

func extractOne(ctx context.Context, extractor domaintext.Extractor, rootPath, inputPath, outputDir string) error {
	mimeType := mimeTypeFromPath(inputPath)
	if mimeType == "" {
		return fmt.Errorf("unsupported file extension")
	}
	output, err := extractor.Extract(ctx, domaintext.ExtractInput{
		LocalPath:        inputPath,
		MimeType:         mimeType,
		OriginalFilename: filepath.Base(inputPath),
	})
	if err != nil {
		return err
	}
	if outputDir == "" {
		fmt.Fprintf(os.Stderr, "method: %s\n", output.Method)
		fmt.Print(output.Text)
		return nil
	}

	targetPath, err := outputTextPath(rootPath, inputPath, outputDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(targetPath, []byte(output.Text), 0o644); err != nil {
		return fmt.Errorf("write text: %w", err)
	}
	fmt.Fprintf(os.Stderr, "ok %s -> %s (%s)\n", inputPath, targetPath, output.Method)
	return nil
}

func outputTextPath(rootPath, inputPath, outputDir string) (string, error) {
	relativePath, err := filepath.Rel(rootPath, inputPath)
	if err != nil || relativePath == "." || strings.HasPrefix(relativePath, "..") {
		relativePath = filepath.Base(inputPath)
	}
	ext := filepath.Ext(relativePath)
	if ext != "" {
		relativePath = strings.TrimSuffix(relativePath, ext)
	}
	return filepath.Join(outputDir, relativePath+".txt"), nil
}

func mimeTypeFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return ""
	}
}

func exitWithError(code int, err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}
