// internal/application/services/textextraction/fallback.go
package textextraction

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/observability"
)

type FallbackExtractor struct {
	primary  domaintext.Extractor
	fallback domaintext.Extractor
}

var _ domaintext.Extractor = (*FallbackExtractor)(nil)

func NewFallbackExtractor(primary, fallback domaintext.Extractor) *FallbackExtractor {
	return &FallbackExtractor{primary: primary, fallback: fallback}
}

func (e *FallbackExtractor) Extract(ctx context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	var failures []error
	for _, attempt := range []struct {
		name      string
		extractor domaintext.Extractor
	}{{"local", e.primary}, {"document_ai", e.fallback}} {
		if ctx.Err() != nil {
			return nil, apperr.Internal("A leitura do documento foi interrompida. O arquivo foi salvo e precisa de revisao.", ctx.Err())
		}
		started := time.Now()
		var output *domaintext.ExtractOutput
		var err error
		if attempt.extractor == nil {
			err = errors.New("extractor unavailable")
		} else {
			output, err = attempt.extractor.Extract(ctx, input)
		}
		if err == nil && (output == nil || !domaintext.IsUsableText(output.Text)) {
			err = errors.New("extracted text is not usable")
		}
		log := observability.FromContext(ctx)
		if err == nil {
			log.Info("exam_ocr_attempt_succeeded", slog.String("provider", attempt.name), slog.Duration("elapsed", time.Since(started)))
			return output, nil
		}
		log.Warn("exam_ocr_attempt_failed", slog.String("provider", attempt.name), slog.Duration("elapsed", time.Since(started)), slog.Any("err", err))
		failures = append(failures, fmt.Errorf("%s OCR: %w", attempt.name, err))
	}
	return nil, apperr.Internal("Nao foi possivel concluir a leitura automatica apos duas tentativas. O arquivo foi salvo e precisa de revisao. Tente novamente mais tarde ou envie o PDF original.", errors.Join(failures...))
}
