<!-- README.md -->
# Sonnda API

API backend da plataforma Sonnda, voltada para atencao primaria a saude e para organizacao do historico clinico centrado no paciente.

A Sonnda resolve um problema recorrente na pratica clinica: pacientes precisam carregar pilhas de exames, perdem documentos e o cuidado fica fragmentado. A proposta e permitir que o paciente armazene e compartilhe seu historico (sem depender de papel/WhatsApp), e que profissionais de saude consigam visualizar e evoluir o paciente com base em um historico longitudinal acessivel via web.

## O que este repositorio entrega (MVP)

- cadastro e gerenciamento de pacientes;
- upload e processamento de exames laboratoriais;
- extracao automatica de dados estruturados via Google Cloud Document AI;
- armazenamento seguro em PostgreSQL (Supabase);
- arquitetura simples por camadas (domain/app/adapters).

> Atencao: este repositorio nao deve conter dados reais de pacientes nem arquivos de configuracao sensiveis (`.env`).

---

## Sumario

- [Sonnda API](#sonnda-api)
  - [O que este repositorio entrega (MVP)](#o-que-este-repositorio-entrega-mvp)
  - [Sumario](#sumario)
  - [Arquitetura](#arquitetura)
  - [Autorizacao (fora do MVP)](#autorizacao-fora-do-mvp)
  - [Stack Tecnologico](#stack-tecnologico)
  - [Logging](#logging)
  - [Endpoints](#endpoints)
  - [Estrutura de Pastas](#estrutura-de-pastas)

---

## Arquitetura

A arquitetura foi simplificada em camadas diretas, com baixo acoplamento:

- **Domain (`internal/domain`)**: modelos de dominio e regras de negocio (agnostico de infraestrutura e HTTP).
- **App (`internal/app`)**: services de aplicacao (orquestracao) e contrato de erros via `internal/app/apperr`.
- **Ports (`internal/domain/ports`)**: interfaces do dominio (integrations e repositories).
- **Adapters (`internal/adapters`)**:
  - **Inbound HTTP (`internal/adapters/inbound/http`)**: rotas, handlers e middlewares.
  - **Outbound (`internal/adapters/outbound`)**: implementacoes concretas (integrations e persistence).

---

## Autorizacao (fora do MVP)

RBAC e ReBAC **foram retirados do MVP**. Nesta fase, o foco e entregar o fluxo principal de cadastro, upload e extracao; a API aplica apenas:

- autenticacao (token Firebase);
- registro no banco (middleware de usuario registrado).

O desenho de autorizacao (RBAC por acoes + ReBAC por relacionamento) fica documentado para uma fase posterior em: `docs/architecture/access-control.md`.

---

## Stack Tecnologico

- **Linguagem:** Go (Golang)
- **Banco de dados:** PostgreSQL (gerenciado via Supabase)
- **ORM / Driver:** `pgx` / `pgxpool`
- **Processamento de documentos:** Google Cloud Document AI
- **Autenticacao:** Firebase Auth (idToken)
- **Containerizacao:** Docker / docker-compose
- **Arquitetura:** camadas simples (domain/app/adapters)
- **Web (UI interna):** `templ` + Tailwind CSS + HTMX (arquivos em `internal/adapters/inbound/http/web`)

---

## Logging

- Logger baseado em `log/slog` (ver `internal/app/observability`).
- Config por env: `LOG_LEVEL` (`debug|info|warn|error`) e `LOG_FORMAT` (`text|json|pretty`).
- Por request, o middleware injeta um logger no `context.Context` (inclui `request_id`, método, path e rota quando disponível).

---

Comandos:

```bash
make dev
```

---

## Endpoints

Indice de rotas expostas em `internal/adapters/inbound/http/api/router.go`.

Publico:
- `GET /api/v1/health`
- `GET /api/v1/docs`

Autenticado (token Firebase):
- `GET /api/v1/check-registration`
- `POST /api/v1/register`

Registrado (token + usuario no banco):
- `GET /api/v1/me`
- `PUT /api/v1/me`
- `POST /api/v1/patients`
- `GET /api/v1/patients`
- `GET /api/v1/patients/:id`
- `GET /api/v1/patients/:id/medical-records/labs`
- `POST /api/v1/patients/:id/medical-records/labs/upload`
- `GET /api/v1/patients/:id/medical-records/labs/summary`

Documentacao complementar:
- `docs/architecture/README.md`
- `docs/architecture/access-control.md`
- `docs/architecture/error-handling.md`
- `docs/dev/setup.md`
- `docs/api/patient.md`

---

## Estrutura de Pastas

Resumo da estrutura:

```text
.
|-- cmd/
|   `-- api/
|       `-- main.go                 # Ponto de entrada da API
|-- internal/
|   |-- app/
|   |   |-- apperr/                 # Contrato de erros da aplicacao
|   |   |-- bootstrap/              # Montagem de dependencias por modulo
|   |   |-- config/                 # Config da aplicacao
|   |   |-- observability/          # Log (slog) e helpers de observability
|   |   `-- services/               # Services de aplicacao (ex.: user, professional, registration)
|   |-- domain/
|   |   |-- model/                  # Modelos e regras de negocio
|   |   `-- ports/                  # Interfaces do dominio (integration, repository)
|   `-- adapters/
|       |-- inbound/http/           # Server HTTP (router + API + WEB)
|       |   |-- api/                # API JSON (handlers, routes, middleware)
|       |   `-- web/                # Web UI (templ + Tailwind + JS)
|       |       |-- assets/static/  # CSS/JS/imagens servidos via /static
|       |       |-- assets/templates/ # Templates .templ (source of truth)
|       |       `-- embed/          # Assets embutidos (ex.: favicon)
|       `-- outbound/               # Integrations e persistence
|           |-- integrations/       # auth, documentai, storage
|           `-- persistence/        # db, repository, sqlc
|-- docker-compose.yml
|-- Dockerfile
|-- Makefile
`-- README.md
```

### Tailwind e templ (WEB)

- **Tailwind**: entrada em `internal/adapters/inbound/http/web/assets/static/css/input.css`, saida em `internal/adapters/inbound/http/web/assets/static/css/app.css` (veja `Makefile`).
- **templ**: templates em `internal/adapters/inbound/http/web/assets/templates/**/*.templ` e arquivos gerados `*_templ.go` (nao edite os `*_templ.go`).
- **Workflow rapido (Windows)**: `make dev-web` (Air + Tailwind watch + templ watch).
