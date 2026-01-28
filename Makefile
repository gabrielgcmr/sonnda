# ==============================================================================
# 🛠️ CONFIGURAÇÕES E VARIÁVEIS
# ==============================================================================
APP_NAME := sonnda
MAIN     := ./cmd/server

# Versões das Ferramentas
AIR_VERSION      := latest
TAILWIND_VERSION := v4.1.18
TEMPL_VERSION    := latest

# Diretórios e Binários
TOOLS_DIR    := tools/bin
AIR          := $(TOOLS_DIR)/air
TAILWIND     := $(TOOLS_DIR)/tailwindcss
TEMPL        := $(TOOLS_DIR)/templ

# Caminhos do Projeto (Preservados do arquivo original)
TAILWIND_INPUT  := internal/adapters/inbound/http/web/styles/input.css
TAILWIND_OUTPUT := internal/adapters/inbound/http/web/static/css/app.css
SQLC_CONF       := internal/adapters/outbound/storage/data/postgres/sqlc/sqlc.yaml

# Detecção de OS/Arch para download dos binários
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
	ARCH := arm64
endif

TAILWIND_ARCH := $(ARCH)
ifeq ($(TAILWIND_ARCH),amd64)
	TAILWIND_ARCH := x64
endif

# Adiciona tools/bin ao PATH para este Makefile
export PATH := $(PWD)/$(TOOLS_DIR):$(PATH)

# ==============================================================================
# 🎯 TARGETS PRINCIPAIS
# ==============================================================================
.PHONY: all dev-api dev-web dev-web-watch build clean test help

all: build

# Instala todas as dependências (Air, Tailwind, Templ)
tools: $(AIR) $(TAILWIND) $(TEMPL)

# Roda apenas o backend (Go + Air)
dev-api: tools
	$(AIR) -c .air.toml

# 🚀 Roda o ambiente COMPLETO (Templ + Tailwind + Air) em paralelo
dev-web: tools
	@echo "🏗️  Gerando assets primeiro..."
	@$(MAKE) templ tailwind 
	@echo "🚀 Subindo servidor..."
	@$(MAKE) air-run     

# 🚀 Roda o ambiente COMPLETO (Templ + Tailwind + Air) em modo -watch
dev-web-watch: tools
	@echo "🔥 Iniciando modo Watch Paralelo... Pode gerar erros de compilação devido a race conditions. Use com sabedoria."
	@$(MAKE) -j3 templ-watch tailwind-watch air-run

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


$(TAILWIND):
	@echo "🎨 Instalando tailwindcss versão: $(TAILWIND_VERSION)..."
	@mkdir -p $(TOOLS_DIR)
	@curl -L -o $(TAILWIND) https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$(OS)-$(TAILWIND_ARCH)
	@chmod +x $(TAILWIND)

$(TEMPL):
	@echo "🔥 Instalando templ versão: $(TEMPL_VERSION)..."
	@GOBIN=$(PWD)/$(TOOLS_DIR) go install github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)


# ==============================================================================
# 🔄 WATCHERS E PROCESSOS INTERNOS
# ==============================================================================
.PHONY: templ templ-watch tailwind-watch air-run

air-run:
	$(AIR) -c .air.toml

templ:
	$(TEMPL) generate

templ-watch:
	$(TEMPL) generate --watch

tailwind:
	$(TAILWIND) -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) 

tailwind-watch:
	$(TAILWIND) -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --watch

# ==============================================================================
# 🐘 DATABASE & DOCKER
# ==============================================================================
.PHONY: sqlc sqlc-check docker-up docker-down

sqlc:
	sqlc generate -f $(SQLC_CONF)

sqlc-check:
	sqlc compile -f $(SQLC_CONF)

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ==============================================================================
# ℹ️ AJUDA
# ==============================================================================
help:
	@echo "Comandos disponíveis:"
	@echo "  dev-api     - Inicia apenas o Backend (Air)"
	@echo "  dev-web     - Inicia Backend + Frontend (Templ/Tailwind) em paralelo"
	@echo "  dev-web-watch - Inicia Backend + Frontend (Templ/Tailwind) em modo watch"
	@echo "  build       - Gera o binário de produção"
	@echo "  tools       - Baixa as ferramentas necessárias (localmente)"
	@echo "  clean       - Limpa pastas geradas"
	@echo "  docker-up   - Sobe o banco de dados"