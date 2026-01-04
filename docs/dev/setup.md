# Setup local (dev)

Guia rapido para subir a API localmente.

## Requisitos
- Go instalado (ver `go.mod`).
- Docker + Docker Compose (opcional, para rodar via container).
- Air (hot reload): https://github.com/air-verse/air
- sqlc (opcional, para gerar queries): https://sqlc.dev

## 1) Variaveis de ambiente
Copie o arquivo `.env.example` para `.env` e ajuste os valores:

```bash
copy .env.example .env
```

Principais chaves:
- `SUPABASE_URL`: string de conexao do PostgreSQL (Supabase).
- `GOOGLE_APPLICATION_CREDENTIALS`: caminho do JSON de credenciais do GCP.
- `GCP_PROJECT_ID`, `GCP_PROJECT_NUMBER`, `GCS_BUCKET`, `GCP_LOCATION`, `DOCAI_LABS_PROCESSOR_ID`.
- `LOG_LEVEL` e `LOG_FORMAT`.

Observacao:
- O `docker-compose.yml` monta `./secrets/sonnda-gcs.json` em `/secrets/sonnda-gcs.json`.
  Se for usar Docker, garanta que o arquivo exista nesse caminho local.

## 2) Rodar localmente (sem Docker)
Opcao simples:

```bash
make run
```

Com hot reload (requer `air`):

```bash
make dev
```

Com hot reload + Tailwind (HTMX):

```bash
make dev-web
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

As migracoes SQL vivem em `internal/infrastructure/persistence/sqlc/sql/migrations/`.

Se voce ja tinha um banco criado antes da mudanca de `users.role` -> `users.account_type`
e da adicao de `professionals.kind`, aplique a migracao:

- `internal/infrastructure/persistence/sqlc/sql/migrations/0004_account_type_and_professional_kind.sql`

## 5) Testes

```bash
make test
```
