// cmd/openapi-validate/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type OpenAPIDoc struct {
	OpenAPI string                 `yaml:"openapi"`
	Info    OpenAPIInfo            `yaml:"info"`
	Paths   map[string]interface{} `yaml:"paths"`
}

type OpenAPIInfo struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "file", "docs/api/openapi.yaml", "path to OpenAPI yaml")
	flag.Parse()

	data, err := os.ReadFile(filePath)
	if err != nil {
		exitWithError(fmt.Errorf("failed to read %s: %w", filePath, err))
	}

	var doc OpenAPIDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		exitWithError(fmt.Errorf("invalid yaml in %s: %w", filePath, err))
	}

	validateDoc(filePath, doc)
	fmt.Printf("openapi ok: %s\n", filePath)
}

func validateDoc(filePath string, doc OpenAPIDoc) {
	if strings.TrimSpace(doc.OpenAPI) == "" {
		exitWithError(fmt.Errorf("missing openapi version in %s", filePath))
	}
	if !strings.HasPrefix(strings.TrimSpace(doc.OpenAPI), "3.") {
		exitWithError(fmt.Errorf("unsupported openapi version %q in %s", doc.OpenAPI, filePath))
	}
	if strings.TrimSpace(doc.Info.Title) == "" {
		exitWithError(fmt.Errorf("missing info.title in %s", filePath))
	}
	if strings.TrimSpace(doc.Info.Version) == "" {
		exitWithError(fmt.Errorf("missing info.version in %s", filePath))
	}
	if len(doc.Paths) == 0 {
		exitWithError(fmt.Errorf("missing paths in %s", filePath))
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
