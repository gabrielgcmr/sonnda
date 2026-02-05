<!-- README.md -->
# Sonnda

Server da plataforma Sonnda, voltada para atencao primaria a saude e para organizacao do historico clinico centrado no paciente.

A Sonnda resolve um problema recorrente na pratica clinica: pacientes precisam carregar pilhas de exames, perdem documentos e o cuidado fica fragmentado. A proposta e permitir que o paciente armazene e compartilhe seu historico (sem depender de papel/WhatsApp), e que profissionais de saude consigam visualizar e evoluir o paciente com base em um historico longitudinal acessivel via web.

## O que este repositorio entrega (MVP)

- cadastro e gerenciamento de pacientes;
- upload e processamento de exames laboratoriais;
- extracao automatica de dados estruturados via Google Cloud Document AI;
- armazenamento seguro em PostgreSQL (Supabase);

> Atencao: este repositorio nao deve conter dados reais de pacientes nem arquivos de configuracao sensiveis (`.env`).

---

## Sumario

- [Sonnda](#sonnda)
  - [O que este repositorio entrega (MVP)](#o-que-este-repositorio-entrega-mvp)
  - [Sumario](#sumario)
  - [Arquitetura](#arquitetura)
  - [Stack Tecnologico](#stack-tecnologico)
  - [Logging](#logging)
  - [Tratamento de Erros](#tratamento-de-erros)
    - [Regras centrais](#regras-centrais)
    - [AppError](#apperror)
  - [Configuracao de ambiente (dev/prod)](#configuracao-de-ambiente-devprod)
  - [Estrutura de Pastas](#estrutura-de-pastas)

---

## Arquitetura

A arquitetura segue uma abordagem em camadas com baixo acoplamento e clara separação de responsabilidades:

- **Domain (`internal/domain`)**: Modelos e regras de negócio centrais (agnóstico de infraestrutura e HTTP).
  - **Entity (`internal/domain/entity`)**: Entidades de negócio centrais.
  - **Model (`internal/domain/model`)**: Estruturas de dados e modelos de domínio.
  - **Repository (`internal/domain/repository`)**: Interfaces de repositórios do domínio.
  - **Ports (`internal/domain/ports`)**: Interfaces do domínio (abstrações para integrações).

- **Application (`internal/application`)**: Orquestração e preocupações transversais.
  - **Use cases (`internal/application/usecase`)**: Fluxos de negócio compostos a partir de modelos/portas do domínio.
  - **Services (`internal/application/services`)**: Serviços de aplicação que coordenam repositórios/integrações.
  - **Config (`internal/application/config`)**: Configuração de ambiente.
  - **Bootstrap (`internal/application/bootstrap`)**: Injeção de dependências e carregamento de configuração.

- **Infrastructure (`internal/infrastructure`)**: Implementações concretas e integrações outbound.
  - **Persistence (`internal/infrastructure/persistence`)**: Repositórios de banco de dados, cache, armazenamento de arquivos.
  - **Auth (`internal/infrastructure/auth`)**: Implementações de autenticação/autorização.

- **Kernel (`internal/kernel`)**: Preocupações transversais.
  - **Error contract (`internal/kernel/apperr`)**: Códigos/mensagens centralizadas de `AppError`; handlers devem converter via helpers da camada HTTP.
  - **Observability (`internal/kernel/observability`)**: Configuração de logging (slog), injeção de logger com escopo de requisição.

---

## Stack Tecnologico

**Backend:** API RESTful em Go
- **Framework web:** Gin
- **Banco de dados:** PostgreSQL (gerenciado via Supabase) + Redis (Upstash)
- **Geração de código SQL:** SQLC
- **Autenticação:** Supabase Auth (JWT)
- **Armazenamento de arquivos:** Supabase Storage
- **Processamento de documentos:** Google Cloud Document AI
- **Containerização:** Docker / docker-compose

**Ferramentas de desenvolvimento:**
- Air (live reload)
- SQLC (geração de código SQL)
- Make (automação de tarefas)
- Docker + docker-compose (containerização)

---

## Logging

- O app usa `log/slog` via `internal/kernel/observability` (logger com escopo de requisição é injetado pelo middleware HTTP).
- Configure com `LOG_LEVEL` (`debug|info|warn|error`) e `LOG_FORMAT` (`text|json|pretty`).

## Tratamento de Erros
Este projeto usa um **contrato de erro centralizado** baseado em `AppError`.

### Regras centrais
- **NÃO retorne strings brutas como contratos de erro.**
- **NÃO exponha `err.Error()` em respostas HTTP.**
- **NÃO construa JSON de erro manualmente em handlers ou middlewares.**

### AppError
- Erros de nível de aplicação devem ser representados como `*apperr.AppError`.
- Localização: `internal/kernel/apperr`
- `AppError` contém:
  - `Code` (`ErrorCode`) - contrato estável e legível por máquina
  - `Message` - mensagem segura e legível por humanos
  - `Cause` - erro interno opcional (wrapped com `%w`)
- **Prefira usar construtores helper** de `internal/kernel/apperr/factory.go` ao invés de construí-los manualmente.
- Services/use cases **devem retornar `AppError` para falhas conhecidas** (validação, conflitos, não encontrado, erros de infra).
- Domínio **nunca** importa HTTP, Gin ou `apperr`.
- Handlers e middlewares **devem chamar**: `httperrors.WriteError(c, err)`
- Apresentação de erro HTTP é centralizada em: `internal/adapters/inbound/http/shared/httperr`.

## Configuracao de ambiente (dev/prod)

- Para dev local, copie o exemplo: `cp .env.example .env`.
- O app **nao** carrega `.env` automaticamente. Exporte as variaveis no shell (ou use `direnv`).

---

Comandos:

```bash
make dev
```

## Estrutura de Pastas

Estrutura atual do projeto:

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # Ponto de entrada da aplicação
├── docs/
│   ├── README.md
│   ├── api/                        # Documentação de endpoints
│   │   ├── auth.md
│   │   ├── labs.md
│   │   ├── openapi.yaml
│   │   ├── patient.md
│   │   ├── user.md
│   │   └── README.md
│   ├── architecture/               # Arquitetura e decisões
│   │   ├── access-control.md
│   │   ├── app-source-of-truth.md
│   │   ├── error-handling.md
│   │   └── adr/                    # Architecture Decision Records
│   └── dev/                        # Guias de desenvolvimento
│       └── setup.md
├── static/                         # Assets embutidos (docs, openapi, favicon)
│   ├── embed.go
│   ├── openapi.yaml               # Gerado a partir de docs/api/openapi.yaml
│   ├── docs.html
│   └── favicon.ico
├── internal/
│   ├── api/                        # [LEGADO - sendo migrado]
│   │   ├── routes.go
│   │   ├── handlers/
│   │   ├── helpers/
│   │   ├── middleware/
│   │   └── presenter/
│   ├── application/                # Camada de aplicação
│   │   ├── bootstrap/              # Injeção de dependências
│   │   ├── services/               # Serviços de aplicação
│   │   └── usecase/                # Casos de uso
│   ├── config/                     # Configuração da aplicação
│   │   ├── config.go
│   │   └── config_test.go
│   ├── domain/                     # Camada de domínio (core)
│   │   ├── ai/                     # Abstrações de IA
│   │   ├── entity/                 # Entidades de domínio
│   │   ├── repository/             # Interfaces de repositório
│   │   └── storage/                # Interfaces de armazenamento
│   ├── infrastructure/             # Implementações concretas
│   │   ├── ai/                     # Integração com Google Cloud Document AI
│   │   ├── auth/                   # Autenticação/autorização
│   │   └── persistence/            # Persistência
│   │       ├── filestorage/        # Armazenamento de arquivos
│   │       ├── postgres/           # Repositórios PostgreSQL
│   │       └── redis/              # Cache Redis
│   └── kernel/                     # Núcleo transversal
│       ├── apperr/                 # Contrato de erros centralizado
│       │   ├── catalog.go
│       │   ├── error.go
│       │   ├── factory.go
│       │   ├── logging.go
│       │   └── violation.go
│       └── observability/          # Logging e observabilidade
│           ├── context.go
│           ├── logger.go
│           ├── pretty_handler.go
│           └── utils.go
├── secrets/                        # Configurações sensíveis (não versionado)
│   └── sonnda-gcs.json
├── tools/                          # Ferramentas de build
│   └── bin/
│       ├── air
│       └── sqlc
├── .env.example                    # Template de variáveis de ambiente
├── AGENTS.md                       # Instruções para agentes de IA
├── docker-compose.yml              # Orquestração de containers
├── Dockerfile                      # Build da imagem Docker
├── go.mod                          # Dependências Go
├── Makefile                        # Comandos úteis
└── README.md                       # Este arquivo
```
