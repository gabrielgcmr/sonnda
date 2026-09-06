// internal/config/ocr_test.go
package config

import (
	"testing"
	"time"
)

func TestOCRTimeoutConfiguration(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  time.Duration
	}{
		{"", time.Minute}, {"90s", 90 * time.Second}, {"0s", 0}, {"-1s", 0}, {"invalid", 0},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("OCR_TIMEOUT", tc.value)
			cfg, err := loadOCRConfig()
			if tc.want == 0 {
				if err == nil {
					t.Fatal("expected invalid configuration")
				}
				return
			}
			if err != nil || cfg.Timeout != tc.want {
				t.Fatalf("got %+v, %v", cfg, err)
			}
		})
	}
}
