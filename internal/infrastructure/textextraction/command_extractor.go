// internal/infrastructure/textextraction/command_extractor.go
package textextraction

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
)

type CommandExtractor struct {
	timeout           time.Duration
	requireUsableText bool
}

type CommandExtractorOptions struct {
	Timeout           time.Duration
	RequireUsableText bool
}

var _ domaintext.Extractor = (*CommandExtractor)(nil)

func NewCommandExtractor() *CommandExtractor {
	return &CommandExtractor{timeout: 20 * time.Second, requireUsableText: true}
}

func NewCommandExtractorWithOptions(options CommandExtractorOptions) *CommandExtractor {
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &CommandExtractor{timeout: timeout, requireUsableText: options.RequireUsableText}
}

func (e *CommandExtractor) Extract(ctx context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
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

func (e *CommandExtractor) extractPDF(ctx context.Context, localPath string) (*domaintext.ExtractOutput, error) {
	text, err := e.run(ctx, "pdftotext", "-raw", localPath, "-")
	if err != nil {
		return nil, err
	}
	if e.requireUsableText && !domaintext.IsUsableText(text) {
		return nil, errors.New("pdf text is not usable")
	}

	return &domaintext.ExtractOutput{
		Text:   text,
		Method: "pdf_text_raw",
	}, nil
}

func (e *CommandExtractor) extractImage(ctx context.Context, localPath string) (*domaintext.ExtractOutput, error) {
	text, method, err := e.extractImageWithRotations(ctx, localPath)
	if err != nil {
		return nil, err
	}
	if e.requireUsableText && !domaintext.IsUsableText(text) {
		return nil, errors.New("ocr text is not usable")
	}

	return &domaintext.ExtractOutput{
		Text:   text,
		Method: method,
	}, nil
}

func (e *CommandExtractor) extractImageWithRotations(ctx context.Context, localPath string) (string, string, error) {
	candidates, cleanup, err := imageRotationCandidates(localPath)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		candidates = []imageCandidate{{path: localPath, rotation: 0}}
	}

	var bestText string
	var bestRotation int
	var bestScore int
	var lastErr error
	for _, candidate := range candidates {
		text, err := e.runTesseract(ctx, candidate.path)
		if err != nil {
			lastErr = err
			continue
		}
		score := scoreOCRText(text)
		if score > bestScore || bestText == "" {
			bestText = text
			bestRotation = candidate.rotation
			bestScore = score
		}
	}
	if bestText == "" {
		if lastErr != nil {
			return "", "", lastErr
		}
		return "", "", errors.New("ocr text is empty")
	}
	method := "ocr"
	if bestRotation != 0 {
		method = fmt.Sprintf("ocr_rotate_%d", bestRotation)
	}
	return bestText, method, nil
}

func (e *CommandExtractor) runTesseract(ctx context.Context, localPath string) (string, error) {
	text, err := e.run(ctx, "tesseract", localPath, "stdout", "-l", "por+eng")
	if err != nil {
		// Alguns ambientes nao possuem os idiomas instalados.
		text, err = e.run(ctx, "tesseract", localPath, "stdout")
	}
	return text, err
}

type imageCandidate struct {
	path     string
	rotation int
}

func imageRotationCandidates(localPath string) ([]imageCandidate, func(), error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, nil, err
	}

	tempDir, err := os.MkdirTemp("", "sonnda-ocr-*")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	candidates := []imageCandidate{{path: localPath, rotation: 0}}
	for _, rotation := range []int{90, 180, 270} {
		targetPath := filepath.Join(tempDir, fmt.Sprintf("rotated-%d.jpg", rotation))
		if err := writeJPEG(targetPath, rotateImage(img, rotation)); err != nil {
			cleanup()
			return nil, nil, err
		}
		candidates = append(candidates, imageCandidate{path: targetPath, rotation: rotation})
	}
	return candidates, cleanup, nil
}

func writeJPEG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return jpeg.Encode(file, img, &jpeg.Options{Quality: 92})
}

func rotateImage(src image.Image, degrees int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	switch degrees {
	case 90:
		dst := image.NewRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(height-1-y, x, src.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
		return dst
	case 180:
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(width-1-x, height-1-y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
		return dst
	case 270:
		dst := image.NewRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dst.Set(y, width-1-x, src.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
		return dst
	default:
		return src
	}
}

func scoreOCRText(text string) int {
	normalized := strings.ToLower(text)
	score := len(strings.Fields(normalized))
	for _, signal := range []string{
		"hemograma", "eritrograma", "leucograma", "plaquetas",
		"paciente", "material", "metodo", "laboratorio", "resultado",
		"hemoglobina", "hematocrito", "leucocitos",
	} {
		if strings.Contains(normalized, signal) {
			score += 80
		}
	}
	for _, r := range normalized {
		switch {
		case r >= '0' && r <= '9':
			score += 1
		case r == '�':
			score -= 10
		}
	}
	return score
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
