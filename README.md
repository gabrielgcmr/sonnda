# Sonnda API

API backend da plataforma Sonnda, voltada para atencao primaria a saude e para organizacao do historico clinico centrado no paciente.

A Sonnda resolve um problema recorrente na pratica clinica: pacientes precisam carregar pilhas de exames, perdem documentos e o cuidado fica fragmentado. A proposta e permitir que o paciente armazene e compartilhe seu historico (sem depender de papel/WhatsApp), e que profissionais de saude consigam visualizar e evoluir o paciente com base em um historico longitudinal acessivel via web.

## O que este repositorio entrega (MVP)

- cadastro e gerenciamento de pacientes;
- upload e processamento de exames laboratoriais;
- extracao automatica de dados estruturados via Google Cloud Document AI;
- armazenamento seguro em PostgreSQL (Supabase);
- arquitetura simples por camadas (domain/app/http/infrastructure).

> Atencao: este repositorio nao deve conter dados reais de pacientes nem arquivos de configuracao sensiveis (`.env`).

---

## Sumario

- [Sonnda API](#sonnda-api)
  - [O que este repositorio entrega (MVP)](#o-que-este-repositorio-entrega-mvp)
  - [Sumario](#sumario)
  - [Arquitetura](#arquitetura)
  - [Stack Tecnologico](#stack-tecnologico)
  - [Endpoints](#endpoints)
  - [Estrutura de Pastas](#estrutura-de-pastas)

---

## Arquitetura

A arquitetura foi simplificada em camadas diretas, com baixo acoplamento:

- **Domain (`internal/domain`)**: entidades e regras de negocio; ports (repositorios e servicos).
- **App (`internal/app`)**: casos de uso e orquestracao; modules conectam dependencias; config e observability.
- **HTTP (`internal/http`)**: API e HTMX com rotas, handlers e middlewares.
- **Infrastructure (`internal/infrastructure`)**: implementacoes concretas para auth, persistence, documentai e storage.

---

## Stack Tecnologico

- **Linguagem:** Go (Golang)
- **Banco de dados:** PostgreSQL (gerenciado via Supabase)
- **ORM / Driver:** `pgx` / `pgxpool`
- **Processamento de documentos:** Google Cloud Document AI
- **Autenticacao:** Firebase Auth (idToken)
- **Containerizacao:** Docker / docker-compose
- **Arquitetura:** camadas simples (domain/app/http/infrastructure)

---

## Logging

- Logger baseado em `log/slog` (ver `internal/app/config/observability`).
- Config por env: `LOG_LEVEL` (`debug|info|warn|error`) e `LOG_FORMAT` (`text|json|pretty`).
- Por request, o middleware injeta um logger no `context.Context` (inclui `request_id`, método, path e rota quando disponível).

---

Comandos:

```bash
make dev-web
```

---

## Endpoints

Indice de rotas expostas em `internal/http/api/router.go`.

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
- `GET /api/v1/patients/medical-records/labs`
- `POST /api/v1/patients/medical-records/labs/upload`
- `GET /api/v1/patients/medical-records/labs/summary`

Documentacao complementar:
- `docs/architecture.md`
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
|-- assets/                         # Assets embedados (ex.: favicon)
|-- internal/
|   |-- app/
|   |   |-- config/                 # Env, observability e config da aplicacao
|   |   |-- modules/                # Montagem de dependencias por modulo
|   |   `-- usecases/               # Casos de uso (orquestracao)
|   |-- domain/
|   |   |-- entities/               # Entidades e regras de negocio
|   |   `-- ports/                  # Interfaces (repositorios/servicos)
|   |-- http/
|   |   |-- api/                    # API HTTP (handlers e rotas)
|   |   |-- htmx/                   # Painel administrativo (HTMX)
|   |   `-- middleware/             # Middlewares HTTP
|   `-- infrastructure/
|       |-- auth/                   # Implementacoes de autenticacao
|       |-- documentai/             # Cliente Google Document AI
|       |-- persistence/            # Supabase + sqlc (repositorios)
|       `-- storage/                # Storage
|-- docker-compose.yml
|-- Dockerfile
|-- Makefile
`-- README.md
```
