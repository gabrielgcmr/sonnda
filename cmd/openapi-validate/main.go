// cmd/openapi-validate/main.go
package main

import (
	"flag"
	"fmt"
	"os"

	openapitools "github.com/gabrielgcmr/sonnda/internal/tooling/openapi"
)

func main() {
	filePath := flag.String("file", "openapi.yaml", "path to OpenAPI entrypoint")
	flag.Parse()
	if _, err := openapitools.Bundle(*filePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("openapi ok: %s\n", *filePath)
}
