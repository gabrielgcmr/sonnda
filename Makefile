# Makefile
# ==============================================================================
# 🛠️ CONFIGURAÇÕES E VARIÁVEIS
# ==============================================================================
APP_NAME := sonnda
MAIN     := ./cmd/api
VERSION ?= 1.0.0
LDFLAGS := -s -w -X github.com/gabrielgcmr/sonnda/cmd/api.version=$(VERSION)
SQLC_SPEC := internal/infrastructure/persistence/postgres/sqlc/sqlc.yaml
OPENAPI_SPEC := internal/api/openapi/openapi.yaml

# ==============================================================================
# 🎯 TARGETS PRINCIPAIS
# ==============================================================================
.PHONY: all dev build clean generate test help openapi-validate oapi-codegen tools-air

all: build

# Roda apenas o backend (Go + Air)
dev:
	@if [ -x "./bin/air" ]; then \
		./bin/air -c .air.toml; \
	elif command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "Aviso: 'air' não encontrado; iniciando sem hot reload."; \
		echo "Para instalar: go install github.com/air-verse/air@latest"; \
		go run $(MAIN); \
	fi

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
	@if [ -x "./bin/air" ]; then \
		./bin/air -c .air.toml; \
	elif command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "Aviso: 'air' não encontrado; iniciando sem hot reload."; \
		echo "Para instalar: go install github.com/air-verse/air@latest"; \
		go run $(MAIN); \
	fi

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
OAPI_CODEGEN_INPUT   := $(OPENAPI_SPEC)
OAPI_CODEGEN_OUTPUT  := internal/api/openapi/generated/oapi.gen.go
OAPI_CODEGEN_PACKAGE := openapi
OAPI_CODEGEN_GENERATE := types,gin

oapi-codegen:
	@mkdir -p $(dir $(OAPI_CODEGEN_OUTPUT))
	go tool oapi-codegen -generate $(OAPI_CODEGEN_GENERATE) -package $(OAPI_CODEGEN_PACKAGE) -o $(OAPI_CODEGEN_OUTPUT) $(OAPI_CODEGEN_INPUT)

generate: sqlc oapi-codegen

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
	@echo "  dev     - Inicia apenas o Backend (Air)"
	@echo "  build       - Gera o binário de produção"
	@echo "  clean       - Limpa pastas geradas"
	@echo "  generate    - Gera código (sqlc + oapi-codegen)"
	@echo "  openapi-validate - Valida o OpenAPI local"
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
