# Makefile
.PHONY: run build dev dev-web tailwind tailwind-watch sqlc-check sqlc docker-up docker-up-build docker-logs docker-down docker-restart clean db-migrate test help

APP_NAME := sonnda-
MAIN     := ./cmd/server
TAILWIND_BIN   := tools/tailwindcss.exe
TAILWIND_INPUT := internal/adapters/inbound/http/web/styles/input.css
TAILWIND_OUTPUT := internal/adapters/inbound/http/web/public/css/app.css

# Executar localmente
run:
	go run $(MAIN)

# Build da aplicação (prod)
build:
	templ generate
	$(TAILWIND_BIN) -c tailwind.config.js -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT)
	go build -o bin/sonnda $(MAIN)

# Executar com hot reload (air)
dev:
	air -c .air.toml

# Templ
templ:
	templ generate --watch

# Tailwind CSS
tailwind:
	$(TAILWIND_BIN) -c tailwind.config.js -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT)

tailwind-watch:
	$(TAILWIND_BIN) -c tailwind.config.js -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --watch

# Hot reload + Tailwind watch (Windows)
dev-web:
	powershell -NoProfile -Command "Start-Process -WindowStyle Hidden -FilePath '$(TAILWIND_BIN)' -ArgumentList '-c','tailwind.config.js','-i','$(TAILWIND_INPUT)','-o','$(TAILWIND_OUTPUT)','--watch'; Start-Process -WindowStyle Hidden -FilePath 'templ' -ArgumentList 'generate','--watch'; air -c .air.toml"

#sqlc
sqlc-check:
	sqlc compile -f internal/infrastructure/persistence/sqlc/sqlc.yaml

sqlc:
	sqlc generate -f internal/adapters/outbound/persistence/sqlc/sqlc.yaml

# Docker
docker-up:
	docker compose up -d

docker-up-build:
	docker compose up -d --build

docker-logs:
	docker logs -f $(APP_NAME)

docker-down:
	docker compose down

docker-restart: docker-down docker-up-build

# Limpeza (Windows-friendly)
clean:
	powershell -NoProfile -Command "if (Test-Path bin) { Remove-Item -Recurse -Force bin }"
	docker system prune -f

# Testes
test:
	go test ./...

# Ajuda
help:
	@echo "Comandos disponíveis:"
	@echo "  run            - Executar aplicação local"
	@echo "  dev            - Executar com hot reload (air)"
	@echo "  dev-web        - Hot reload + Tailwind watch + templ watch"
	@echo "  tailwind       - Build do CSS Tailwind"
	@echo "  tailwind-watch - Watch do CSS Tailwind"
	@echo "  templ          - Watch do templ"
	@echo "  build          - Build (templ + tailwind + go)"
	@echo "  sqlc-check     - Validar queries SQLC"
	@echo "  sqlc           - Gerar código SQLC"
	@echo "  docker-up      - Subir containers (background)"
	@echo "  docker-up-build- Subir containers com rebuild"
	@echo "  docker-logs    - Ver logs do container"
	@echo "  docker-down    - Parar containers"
	@echo "  docker-restart - Reiniciar containers (rebuild)"
	@echo "  test           - Executar testes"
