<!-- docs/dev/setup.md -->
# Setup local (dev)

Guia rapido para subir a API localmente.

## Requisitos
- Go instalado (ver `go.mod`).
- Instale make via:
  - sudo apt-get update
  - sudo apt-get install make
- Rode o comando: `make tools`

## 1) Variaveis de ambiente
Copie o exemplo para `.env`:

```bash
cp .env.example .env
```

Observacao:
- O app nao carrega `.env` automaticamente. Exporte as variaveis no shell (ou use `direnv`).
- O `docker-compose.yml` monta `./secrets/sonnda-gcs.json` em `/secrets/sonnda-gcs.json`.
  Se for usar Docker, garanta que o arquivo exista nesse caminho local.

## 2) Rodar localmente (sem Docker)
Opcao simples:

```bash
set -a && source .env && set +a
go run ./cmd/server
```

Com hot reload (requer `air`):

```bash
make dev-api
```

## 3) Rodar via Docker

```bash
make docker-up
```

Para rebuild:

```bash
make docker-up-build
```

## 4) sqlc (quando alterar SQL)

```bash
make sqlc-check
make sqlc-generate
```

## Migracoes

As migracoes SQL vivem em `internal/adapters/outbound/storage/postgres/sqlc/sql/migrations/`.

## 5) Testes

```bash
make test
```
