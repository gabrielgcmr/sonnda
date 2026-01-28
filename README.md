<!-- README.md -->
# Sonnda

Server da plataforma Sonnda, voltada para atencao primaria a saude e para organizacao do historico clinico centrado no paciente.

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

- [Sonnda](#sonnda)
  - [O que este repositorio entrega (MVP)](#o-que-este-repositorio-entrega-mvp)
  - [Sumario](#sumario)
  - [Arquitetura](#arquitetura)
  - [Stack Tecnologico](#stack-tecnologico)
  - [Logging](#logging)
  - [Endpoints](#endpoints)
  - [Estrutura de Pastas](#estrutura-de-pastas)
    - [Layers (Arquitetura em Camadas)](#layers-arquitetura-em-camadas)
    - [Web (HTMX + Templ + Tailwind CSS)](#web-htmx--templ--tailwind-css)
    - [Fluxo de Build (Web)](#fluxo-de-build-web)

---

## Arquitetura

A arquitetura foi simplificada em camadas diretas, com baixo acoplamento:

- **Domain (`internal/domain`)**: modelos de dominio e regras de negocio (agnostico de infraestrutura e HTTP).
- **App (`internal/app`)**: services de aplicacao (orquestracao) e contrato de erros via `internal/shared/apperr`.
- **Adapters (`internal/adapters`)**:
  - **Inbound (`internal/adapters/inbound/http`)**: protocolo http
    - **Api (`internal/adapters/inbound/http/api`)**: rotas, handlers e middlewares para API.
    - **Web (`internal/adapters/inbound/http/web`)**: rotas, handlers e middlewares para WEB.
  - **Outbound (`internal/adapters/outbound`)**: implementacoes concretas (integrations e persistence).

---

## Stack Tecnologico

- **Linguagem:** Go (Golang)
- **Banco de dados:** PostgreSQL (gerenciado via Supabase)
- **ORM / Driver:** `pgx` / `pgxpool`
- **Processamento de documentos:** Google Cloud Document AI
- **Autenticacao:** Firebase Auth (idToken) cookie
- **Containerizacao:** Docker / docker-compose
- **Arquitetura:** camadas simples (domain/app/adapters)
- **Web (UI interna):** `templ` + Tailwind CSS + HTMX (arquivos em `internal/adapters/inbound/http/web`)

---

## Logging

- Logger baseado em `log/slog` (ver `internal/shared/observability`).
- Config por env: `LOG_LEVEL` (`debug|info|warn|error`) e `LOG_FORMAT` (`text|json|pretty`).
- Por request, o middleware injeta um logger no `context.Context` (inclui `request_id`, método, path e rota quando disponível).

## Configuracao de ambiente (dev/prod)

- Para dev local, copie o exemplo: `cp .env.example .env`.
- O app **nao** carrega `.env` automaticamente. Exporte as variaveis no shell (ou use `direnv`).
- A aplicacao roteia por host: em dev use `app.localhost` e `api.localhost` (caso nao resolva, adicione no `/etc/hosts`).

---

Comandos:

```bash
make dev-api
```

---
<!-- TODO: Remover ending points e usar openapi -->
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

Resumo da estrutura atual:

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # Ponto de entrada da aplicação
├── docs/
│   ├── README.md
│   ├── api/                        # Documentação de endpoints
│   ├── architecture/               # Arquitetura e decisões
│   │   └── adr/                    # Architecture Decision Records
│   └── dev/                        # Guias de desenvolvimento
│       └── setup.md
├── internal/
│   ├── adapters/                   # Adaptadores (inbound/outbound)
│   │   ├── inbound/
│   │   │   ├── cli/                # CLI adapter (futuro)
│   │   │   └── http/               # HTTP adapter (API + WEB)
│   │   │       ├── api/            # API JSON
│   │   │       ├── web/            # Web UI (Templ + Tailwind + HTMX)
│   │   │       │   ├── handlers/   # Web handlers
│   │   │       │   ├── middleware/ # Web middleware
│   │   │       │   ├── static/     # Static assets (CSS, JS, images)
│   │   │       │   ├── styles/     # Tailwind config, theme, fonts
│   │   │       │   └── templates/  # Templ components e páginas
│   │   │       └── router.go
│   │   └── outbound/               # Integrações externas
│   ├── app/                        # Camada de aplicação
│   │   ├── apperr/                 # Contrato de erros
│   │   ├── bootstrap/              # Injeção de dependências
│   │   ├── config/                 # Configuração da aplicação
│   │   ├── observability/          # Logging e observabilidade
│   │   ├── services/               # Services de aplicação
│   │   └── usecase/                # Casos de uso
│   └── domain/                     # Camada de domínio (core)
│       ├── model/                  # Modelos e regras de negócio
│       └── ports/                  # Interfaces do domínio
├── samples/                        # Exemplos e dados de teste
├── secrets/                        # Configurações sensíveis (não versionado)
│   └── sonnda-gcs.json/
├── tools/                          # Ferramentas de build
├── .env.example                    # Template de variáveis de ambiente
├── docker-compose.yml              # Orquestração de containers
├── Dockerfile                      # Build da imagem Docker
├── Makefile                        # Comandos úteis
├── tailwind.config.js              # Configuração do Tailwind CSS
└── README.md                       # Este arquivo
```

### Layers (Arquitetura em Camadas)

A aplicação segue a separação clara de responsabilidades:

1. **Domain (`internal/domain`)**: modelos puros e regras de negócio (independente de infraestrutura/HTTP)
2. **App (`internal/app`)**: services de aplicação, orquestração e contrato de erros
3. **Adapters (`internal/adapters`)**: integração com HTTP, banco de dados e serviços externos

### Web (HTMX + Templ + Tailwind CSS)

- **Source of truth**: `internal/adapters/inbound/http/web/`
- **Componentes**: `templates/components/` (UI reusáveis em `.templ`)
- **Páginas**: `templates/pages/` (páginas completas)
- **Styles**: 
  - `styles/input.css` (entrypoint do Tailwind)
  - `styles/theme.css` (design tokens e cores Material Design 3)
  - `static/css/app.css` (output compilado, não editar)
- **Assets estáticos**: `static/` (favicon, imagens, JS)

### Fluxo de Build (Web)

```bash
# Watch mode (desenvolvimento rápido)
make dev-web        # Tailwind + Templ + Air

# Build one-time
make tailwind       # Compila CSS
make templ          # Gera Go code a partir dos .templ
```
