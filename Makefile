# Makefile
.PHONY: run build dev dev-web sqlc-check sqlc-generate docker-up docker-up-build docker-logs docker-down docker-restart clean db-migrate test help

APP_NAME := sonnda-api``
MAIN     := ./cmd/api

# Executar localmente
run:
	go run $(MAIN)

# Build da aplicação
build:
	go build -o bin/api $(MAIN)

# Executar com hot reload (air)
dev:
	air -c .air.toml

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

# Limpeza
clean:
	rm -rf bin/
	docker system prune -f

# Testes
test:
	go test ./...

# Ajuda
help:
	@echo "Comandos disponíveis:"
	@echo "  run            - Executar aplicação local"
	@echo "  dev            - Executar com hot reload (air)"
	@echo "  build          - Build da aplicação"
	@echo "  sqlc-check     - Validar queries SQLC"
	@echo "  sqlc-generate  - Gerar código SQLC"
	@echo "  docker-up      - Subir containers (background)"
	@echo "  docker-up-build- Subir containers com rebuild"
	@echo "  docker-logs    - Ver logs do container"
	@echo "  docker-down    - Parar containers"
	@echo "  docker-restart - Reiniciar containers (rebuild)"
	@echo "  db-migrate     - Rodar migrations"
	@echo "  test           - Executar testes"
