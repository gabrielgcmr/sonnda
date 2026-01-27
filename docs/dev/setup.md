<!-- docs/dev/setup.md -->
# Setup local (dev)

Guia rapido para subir a API localmente.

## Requisitos
- Go instalado (ver `go.mod`).
- Instale make via:
  - sudo apt-get update
  - sudo apt-get install make
- Rode o comando: make tools
- OU baixe o standalone do Air e Tailwind para a pasta /tools/bin
- templ (templates do WEB): `go install github.com/a-h/templ/cmd/templ@v0.3.977`

## 1) Variaveis de ambiente
Copie o arquivo `.env.example` para `.env` e ajuste os valores:

```bash
copy .env.example .env
```

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

Com hot reload + Tailwind + templ (WEB):

```bash
make dev-web
```

Observacoes (WEB):
- Tailwind usa `tools/tailwindcss` 
- Gera `internal/adapters/inbound/http/web/static/css/app.css` a partir de `internal/adapters/inbound/http/web/styles/input.css` (gerado, nao editar o output).
- `templ` usa os arquivos `.templ` em `internal/adapters/inbound/http/web/templates/` e gera os `*_templ.go` (arquivos gerados, nao editar).

Alternativas (mais controle):
- **Simples**: `make dev-web`
- **Correto (watchers)**: `templ generate --watch` em um terminal e `make tailwind-watch` em outro, junto com `make dev`.

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
