// cmd/openapi-bundle/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	openapitools "github.com/gabrielgcmr/sonnda/internal/tooling/openapi"
)

func main() {
	input := flag.String("input", "openapi.yaml", "OpenAPI source entrypoint")
	output := flag.String("output", "bin/openapi.yaml", "standalone bundle output")
	flag.Parse()
	data, err := openapitools.Bundle(*input)
	if err == nil {
		err = os.MkdirAll(filepath.Dir(*output), 0o755)
	}
	if err == nil {
		err = os.WriteFile(*output, data, 0o644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
