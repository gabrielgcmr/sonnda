# Makefile
# ==============================================================================
# 🛠️ CONFIGURAÇÕES E VARIÁVEIS
# ==============================================================================
APP_NAME := sonnda
MAIN     := ./cmd/server
VERSION ?= 1.0.0
LDFLAGS := -s -w -X github.com/gabrielgcmr/sonnda/internal/api.rootAPIVersion=$(VERSION)

# Versões das Ferramentas
AIR_VERSION      := latest
SQLC_VERSION     := latest

# Diretórios e Binários
TOOLS_DIR    := tools/bin
AIR          := $(TOOLS_DIR)/air
SQLC         := $(TOOLS_DIR)/sqlc

# Caminhos do Projeto (Preservados do arquivo original)
SQLC_CONF  := internal/adapters/outbound/storage/data/postgres/sqlc/sqlc.yaml

# Detecção de OS/Arch para download dos binários
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
	ARCH := arm64
endif

# Adiciona tools/bin ao PATH para este Makefile
export PATH := $(PWD)/$(TOOLS_DIR):$(PATH)

# ==============================================================================
# 🎯 TARGETS PRINCIPAIS
# ==============================================================================
.PHONY: all dev build clean test help sync-openapi sync-redoc openapi-validate

all: build

# Instala todas as dependências (Air, SQLC)
tools: $(AIR) $(SQLC)

# Roda apenas o backend (Go + Air)
dev: tools
	$(AIR) -c .air.toml

build: sync-openapi
	go build -o bin/$(APP_NAME) -ldflags "$(LDFLAGS)" $(MAIN)

# Limpeza (Compatível com Linux/WSL)
clean:
	@echo "🧹 Limpando binários e cache..."
	rm -rf bin $(TOOLS_DIR)

test:
	go test ./... -v

# ==============================================================================
# 📦 INSTALAÇÃO DE FERRAMENTAS (Auto-Download)
# ==============================================================================
$(AIR):
	@echo "☁️  Instalando air versão: $(AIR_VERSION)..."
	@mkdir -p $(TOOLS_DIR)
	@GOBIN=$(PWD)/$(TOOLS_DIR) go install github.com/air-verse/air@$(AIR_VERSION)

$(SQLC):
	@echo "🗄️  Instalando sqlc versão: $(SQLC_VERSION)..."
	@GOBIN=$(PWD)/$(TOOLS_DIR) go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

# ==============================================================================
# 🔄 WATCHERS E PROCESSOS INTERNOS
# ==============================================================================
.PHONY: air-run

air-run:
	$(AIR) -c .air.toml

# ==============================================================================
# 🐘 DATABASE
# ==============================================================================
.PHONY: sqlc sqlc-check 

sqlc: $(SQLC)
	$(SQLC) generate -f $(SQLC_CONF)

sqlc-check: $(SQLC)
	$(SQLC) compile -f $(SQLC_CONF)

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
	@echo "  tools       - Baixa as ferramentas necessárias (localmente)"
	@echo "  clean       - Limpa pastas geradas"
	@echo "  sync-openapi - Sincroniza o OpenAPI em assets"
	@echo "  sync-redoc  - Baixa o bundle do Redoc para assets"
	@echo "  openapi-validate - Valida o OpenAPI local"
	@echo "  docker-up   - Sobe o docker"
	@echo "  docker-down - Derruba o docker"

# ==============================================================================
# 📚 OPENAPI
# ==============================================================================
OPENAPI_DOCS := docs/api/openapi.yaml
OPENAPI_ASSETS := internal/api/assets/openapi.yaml

sync-openapi:
	@mkdir -p $(dir $(OPENAPI_ASSETS))
	@{ \
		echo "# internal/api/assets/openapi.yaml"; \
		echo "# NOTE: keep in sync with docs/api/openapi.yaml"; \
		tail -n +3 $(OPENAPI_DOCS); \
	} > $(OPENAPI_ASSETS)

OPENAPI_VALIDATE := ./cmd/openapi-validate

openapi-validate:
	go run $(OPENAPI_VALIDATE) -file $(OPENAPI_DOCS)

REDOC_URL := https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js
REDOC_ASSETS := internal/api/assets/redoc.standalone.js

sync-redoc:
	@mkdir -p $(dir $(REDOC_ASSETS))
	@{ \
		echo "// internal/api/assets/redoc.standalone.js"; \
		curl -fsSL $(REDOC_URL); \
	} > $(REDOC_ASSETS)
