// internal/application/services/textextraction/fallback_test.go
package textextraction

import (
	"context"
	"errors"
	"strings"
	"testing"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

const usableText = "GLICOSE Resultado: 90 mg/dL. Valor de referencia: 70 a 99 mg/dL. Material: sangue. Metodo: Hexoquinase."

type fakeExtractor struct {
	output *domaintext.ExtractOutput
	err    error
	calls  int
	input  domaintext.ExtractInput
}

func (f *fakeExtractor) Extract(_ context.Context, input domaintext.ExtractInput) (*domaintext.ExtractOutput, error) {
	f.calls++
	f.input = input
	return f.output, f.err
}

func TestFallbackRecoversFailedOrUnusableLocalOCR(t *testing.T) {
	for _, tc := range []struct {
		name   string
		output *domaintext.ExtractOutput
		err    error
	}{
		{"failure", nil, errors.New("tesseract missing")},
		{"local timeout", nil, context.DeadlineExceeded},
		{"nil output", nil, nil},
		{"empty text", &domaintext.ExtractOutput{}, nil},
		{"noise", &domaintext.ExtractOutput{Text: "12345"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			primary := &fakeExtractor{output: tc.output, err: tc.err}
			fallback := &fakeExtractor{output: &domaintext.ExtractOutput{Text: usableText, Method: "document_ai_ocr"}}
			input := domaintext.ExtractInput{DocumentURI: "gs://bucket/photo.jpg", MimeType: "image/jpeg"}
			got, err := NewFallbackExtractor(primary, fallback).Extract(context.Background(), input)
			if err != nil || got != fallback.output || fallback.calls != 1 || fallback.input != input {
				t.Fatalf("failed recovery: output=%+v err=%v calls=%d", got, err, fallback.calls)
			}
		})
	}
}

func TestUsableLocalOCRSkipsCloud(t *testing.T) {
	primary := &fakeExtractor{output: &domaintext.ExtractOutput{Text: usableText, Method: "ocr_psm_6_partial"}}
	fallback := &fakeExtractor{}
	got, err := NewFallbackExtractor(primary, fallback).Extract(context.Background(), domaintext.ExtractInput{})
	if err != nil || got != primary.output || fallback.calls != 0 {
		t.Fatal("usable local OCR must avoid cloud call")
	}
}

func TestFailedFallbackKeepsInternalCausesAndSafeMessage(t *testing.T) {
	localErr := errors.New("private local detail")
	cloudErr := errors.New("private cloud detail")
	_, err := NewFallbackExtractor(&fakeExtractor{err: localErr}, &fakeExtractor{err: cloudErr}).Extract(context.Background(), domaintext.ExtractInput{})
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) || !errors.Is(err, localErr) || !errors.Is(err, cloudErr) {
		t.Fatalf("missing error chain: %v", err)
	}
	if strings.Contains(appErr.Message, "private") || !strings.Contains(appErr.Message, "duas tentativas") {
		t.Fatalf("unsafe or unhelpful message: %s", appErr.Message)
	}
}

func TestCanceledRequestSkipsBothProviders(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	primary, fallback := &fakeExtractor{}, &fakeExtractor{}
	_, err := NewFallbackExtractor(primary, fallback).Extract(ctx, domaintext.ExtractInput{})
	if !errors.Is(err, context.Canceled) || primary.calls != 0 || fallback.calls != 0 {
		t.Fatal("canceled request performed OCR")
	}
}

func TestUnusableCloudTextRemainsAnError(t *testing.T) {
	_, err := NewFallbackExtractor(nil, &fakeExtractor{output: &domaintext.ExtractOutput{Text: "noise"}}).Extract(context.Background(), domaintext.ExtractInput{})
	if err == nil {
		t.Fatal("unusable cloud output must not be accepted")
	}
}
