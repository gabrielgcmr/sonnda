# Makefile
# ==============================================================================
# 🛠️ CONFIGURAÇÕES E VARIÁVEIS
# ==============================================================================
APP_NAME := sonnda
MAIN     := ./cmd/server

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
.PHONY: all api build clean test help

all: build

# Instala todas as dependências (Air, SQLC)
tools: $(AIR) $(SQLC)

# Roda apenas o backend (Go + Air)
api: tools
	$(AIR) -c .air.toml

build:
	go build -o bin/$(APP_NAME) $(MAIN)

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
	@echo "  api     - Inicia apenas o Backend (Air)"
	@echo "  build       - Gera o binário de produção"
	@echo "  tools       - Baixa as ferramentas necessárias (localmente)"
	@echo "  clean       - Limpa pastas geradas"
	@echo "  docker-up   - Sobe o docker"
	@echo "  docker-down - Derruba o docker"