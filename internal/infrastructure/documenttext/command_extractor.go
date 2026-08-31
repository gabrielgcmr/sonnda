package documenttext

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	domaindoc "github.com/gabrielgcmr/sonnda/internal/domain/documenttext"
)

type CommandExtractor struct {
	timeout time.Duration
}

var _ domaindoc.Extractor = (*CommandExtractor)(nil)

func NewCommandExtractor() *CommandExtractor {
	return &CommandExtractor{timeout: 20 * time.Second}
}

func (e *CommandExtractor) Extract(ctx context.Context, input domaindoc.ExtractInput) (*domaindoc.ExtractOutput, error) {
	if strings.TrimSpace(input.LocalPath) == "" {
		return nil, errors.New("local path is required")
	}

	switch normalizeMimeType(input.MimeType) {
	case "application/pdf", "image/pdf":
		return e.extractPDF(ctx, input.LocalPath)
	case "image/jpeg", "image/jpg", "image/png":
		return e.extractImage(ctx, input.LocalPath)
	default:
		return nil, fmt.Errorf("unsupported mime type: %s", input.MimeType)
	}
}

func (e *CommandExtractor) extractPDF(ctx context.Context, localPath string) (*domaindoc.ExtractOutput, error) {
	text, err := e.run(ctx, "pdftotext", localPath, "-")
	if err != nil {
		return nil, err
	}
	if !domaindoc.IsUsableText(text) {
		return nil, errors.New("pdf text is not usable")
	}

	return &domaindoc.ExtractOutput{
		Text:   text,
		Method: "pdf_text",
	}, nil
}

func (e *CommandExtractor) extractImage(ctx context.Context, localPath string) (*domaindoc.ExtractOutput, error) {
	text, err := e.run(ctx, "tesseract", localPath, "stdout", "-l", "por+eng")
	if err != nil {
		// Alguns ambientes nao possuem os idiomas instalados.
		text, err = e.run(ctx, "tesseract", localPath, "stdout")
	}
	if err != nil {
		return nil, err
	}
	if !domaindoc.IsUsableText(text) {
		return nil, errors.New("ocr text is not usable")
	}

	return &domaindoc.ExtractOutput{
		Text:   text,
		Method: "ocr",
	}, nil
}

func (e *CommandExtractor) run(ctx context.Context, name string, args ...string) (string, error) {
	if _, err := exec.LookPath(name); err != nil {
		return "", err
	}

	timeout := e.timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, name, args...)
	output, err := cmd.Output()
	if runCtx.Err() != nil {
		return "", runCtx.Err()
	}
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func normalizeMimeType(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.Index(normalized, ";"); idx >= 0 {
		normalized = strings.TrimSpace(normalized[:idx])
	}
	return normalized
}
