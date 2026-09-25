# Makefile
# ==============================================================================
# 🛠️ CONFIGURAÇÕES E VARIÁVEIS
# ==============================================================================
APP_NAME := sonnda
MAIN     := ./cmd/api
VERSION ?= 1.0.0
LDFLAGS := -s -w -X github.com/gabrielgcmr/sonnda/cmd/api.version=$(VERSION)
SQLC_SPEC := internal/infrastructure/persistence/postgres/sqlc/sqlc.yaml
OPENAPI_SPEC := openapi.yaml
OPENAPI_BUNDLE := dist/openapi.yaml

# ==============================================================================
# 🎯 TARGETS PRINCIPAIS
# ==============================================================================
.PHONY: all dev dev-air build clean generate test help openapi-validate openapi-bundle openapi-validate-bundle openapi-embed openapi-generate oapi-codegen tools-air

all: build

# Roda apenas o backend (sem Air)
dev:
	go run $(MAIN)

# Roda backend com hot reload via Air
dev-air:
	go run github.com/air-verse/air@latest -c .air.toml

build:
	go build -o bin/$(APP_NAME) -ldflags "$(LDFLAGS)" $(MAIN)

# Limpeza (Compatível com Linux/WSL)
clean:
	@echo "🧹 Limpando binários e cache..."
	rm -rf bin

test:
	go test ./... -v

# ==============================================================================
# 🔄 WATCHERS E PROCESSOS INTERNOS
# ==============================================================================
.PHONY: air-run

air-run:
	go run github.com/air-verse/air@latest -c .air.toml

# Instala o Air localmente em ./bin (para uso em dev / CI)
tools-air:
	@mkdir -p bin
	GOBIN=$(CURDIR)/bin go install github.com/air-verse/air@latest

# ==============================================================================
# 🐘 DATABASE
# ==============================================================================
.PHONY: sqlc sqlc-check 

sqlc:
	go tool sqlc generate -f $(SQLC_SPEC)

sqlc-check:
	go tool sqlc compile -f $(SQLC_SPEC)

# ==============================================================================
# 🧬 CODEGEN
# ==============================================================================
OAPI_CODEGEN_INPUT   := $(OPENAPI_BUNDLE)
OAPI_CODEGEN_OUTPUT  := internal/generated/openapi/oapi.gen.go
OAPI_CODEGEN_PACKAGE := openapi
OAPI_CODEGEN_GENERATE := types,gin

openapi-bundle: openapi-validate
	go run ./cmd/openapi-bundle -input $(OPENAPI_SPEC) -output $(OPENAPI_BUNDLE)

openapi-validate-bundle: openapi-bundle
	go run ./cmd/openapi-validate -file $(OPENAPI_BUNDLE)

openapi-embed: openapi-validate-bundle
	go generate ./internal/openapispec

oapi-codegen: openapi-validate-bundle
	go tool oapi-codegen -generate $(OAPI_CODEGEN_GENERATE) -package $(OAPI_CODEGEN_PACKAGE) -o $(OAPI_CODEGEN_OUTPUT) $(OAPI_CODEGEN_INPUT)

openapi-generate: openapi-embed oapi-codegen

generate: sqlc openapi-generate

# ==============================================================================
# 🐘 DOCKER
# ==============================================================================
.PHONY: docker-up docker-down

docker-up:
	docker compose up -d

docker-down:
	docker compose down	

# ==============================================================================
# ℹ️ AJUDA
# ==============================================================================
help:
	@echo "Comandos disponíveis:"
	@echo "  dev         - Inicia apenas o Backend (sem Air)"
	@echo "  dev-air     - Inicia apenas o Backend (com Air)"
	@echo "  build       - Gera o binário de produção"
	@echo "  clean       - Limpa pastas geradas"
	@echo "  generate    - Gera SQLC, bundle OpenAPI, embed e tipos Go"
	@echo "  openapi-generate - Valida, cria o bundle, embed e tipos Go"
	@echo "  openapi-embed - Gera o asset Go a partir do bundle"
	@echo "  openapi-validate - Valida o OpenAPI local"
	@echo "  openapi-bundle - Unifica o OpenAPI em dist/openapi.yaml"
	@echo "  openapi-validate-bundle - Valida dist/openapi.yaml"
	@echo "  tools-air   - Instala o Air em ./bin"
	@echo "  docker-up   - Sobe o docker"
	@echo "  docker-down - Derruba o docker"

# ==============================================================================
# 📚 OPENAPI
# ==============================================================================
OPENAPI_DOCS := $(OPENAPI_SPEC)

OPENAPI_VALIDATE := ./cmd/openapi-validate

openapi-validate:
	go run $(OPENAPI_VALIDATE) -file $(OPENAPI_DOCS)
