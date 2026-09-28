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
	documentaiinfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/documentai"
	textextractioninfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/textextraction"
	"github.com/joho/godotenv"
)

func main() {
	var inputPath string
	var outputDir string
	var requireUsable bool
	var useDocumentAI bool
	flag.StringVar(&inputPath, "input", "", "file or directory with PDFs/images")
	flag.StringVar(&outputDir, "output", "", "directory for generated .txt files; defaults to stdout")
	flag.BoolVar(&requireUsable, "require-usable", false, "fail when extracted text does not pass the runtime quality filter")
	flag.BoolVar(&useDocumentAI, "document-ai", false, "extract text with Google Document AI using the local file contents")
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
	var documentAIClient *documentaiinfra.Client
	var documentAIProcessorID string
	if useDocumentAI {
		_ = godotenv.Load()
		projectID := strings.TrimSpace(os.Getenv("GCP_PROJECT_ID"))
		location := strings.TrimSpace(os.Getenv("GCP_LOCATION"))
		documentAIProcessorID = strings.TrimSpace(os.Getenv("GCP_EXTRACT_LABS_PROCESSOR_ID"))
		if projectID == "" || location == "" || documentAIProcessorID == "" {
			exitWithError(2, errors.New("Document AI requires GCP_PROJECT_ID, GCP_LOCATION, and GCP_EXTRACT_LABS_PROCESSOR_ID"))
		}
		var err error
		documentAIClient, err = documentaiinfra.NewClient(ctx, projectID, location)
		if err != nil {
			exitWithError(1, fmt.Errorf("create Document AI client: %w", err))
		}
		defer documentAIClient.Close()
	}

	if !info.IsDir() {
		if err := extractOne(ctx, extractor, documentAIClient, documentAIProcessorID, inputPath, inputPath, outputDir); err != nil {
			exitWithError(1, err)
		}
		return
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
		if err := extractOne(ctx, extractor, documentAIClient, documentAIProcessorID, inputPath, path, outputDir); err != nil {
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

func extractOne(
	ctx context.Context,
	extractor domaintext.Extractor,
	documentAIClient *documentaiinfra.Client,
	documentAIProcessorID string,
	rootPath, inputPath, outputDir string,
) error {
	mimeType := mimeTypeFromPath(inputPath)
	if mimeType == "" {
		return fmt.Errorf("unsupported file extension")
	}
	var output *domaintext.ExtractOutput
	var err error
	if documentAIClient != nil {
		content, readErr := os.ReadFile(inputPath)
		if readErr != nil {
			return fmt.Errorf("read input: %w", readErr)
		}
		document, processErr := documentAIClient.ProcessRawDocument(ctx, documentAIProcessorID, content, mimeType)
		if processErr != nil {
			return processErr
		}
		output = &domaintext.ExtractOutput{
			Text:           document.GetText(),
			NormalizedText: domaintext.NormalizeForSemanticExtraction(document.GetText()),
			Method:         "document_ai_ocr",
		}
		if strings.TrimSpace(output.Text) == "" {
			return fmt.Errorf("Document AI returned no text")
		}
	} else {
		output, err = extractor.Extract(ctx, domaintext.ExtractInput{
			LocalPath:        inputPath,
			MimeType:         mimeType,
			OriginalFilename: filepath.Base(inputPath),
		})
	}
	if err != nil {
		return err
	}
	if outputDir == "" {
		fmt.Fprintf(os.Stderr, "extracted %s (%s)\n", inputPath, output.Method)
		fmt.Printf("===== %s =====\n%s\n", filepath.Base(inputPath), output.Text)
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
