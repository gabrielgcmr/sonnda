// cmd/openapi-export/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gabrielgcmr/sonnda/internal/api"
)

func main() {
	outputPath := flag.String("output", "artifacts/openapi.json", "OpenAPI JSON output path")
	name := flag.String("name", "Sonnda API", "API name")
	version := flag.String("version", "dev", "API revision or version")
	environment := flag.String("environment", "ci", "API environment")
	flag.Parse()

	spec := api.OpenAPI(api.RootInfo{Name: *name, Version: *version, Env: *environment})
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fail(fmt.Errorf("marshal OpenAPI: %w", err))
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(fmt.Errorf("create output directory: %w", err))
	}
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fail(fmt.Errorf("write OpenAPI: %w", err))
	}
	fmt.Printf("OpenAPI exported to %s\n", *outputPath)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
