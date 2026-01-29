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
  - [Configuracao de ambiente (dev/prod)](#configuracao-de-ambiente-devprod)
  - [Estrutura de Pastas](#estrutura-de-pastas)

---

## Arquitetura

A arquitetura foi simplificada em camadas diretas, com baixo acoplamento:

- **Domain (`internal/domain`)**: modelos de dominio e regras de negocio (agnostico de infraestrutura e HTTP).
- **App (`internal/app`)**: services de aplicacao (orquestracao)
- **Adapters (`internal/adapters`)**:
  - **Inbound (`internal/adapters/inbound/http`)**: protocolo http (e grpc, cli no futuro)
    - **Api (`internal/adapters/inbound/http/api`)**: rotas, handlers e middlewares para API.
  - **Outbound (`internal/adapters/outbound`)**: implementacoes concretas (integrations e persistence).
- **Kernel (`internal/kernel`)**: Núcleo transversal do sistema. Com contrato de erros (apperr) e log (observability)

---

## Stack Tecnologico

- **Linguagem:** Go (Golang)
- **Banco de dados:** PostgreSQL (gerenciado via Supabase)
- **ORM / Driver:** `pgx` / `pgxpool`
- **Processamento de documentos:** Google Cloud Document AI
- **Autenticacao:** Supabase Auth (JWT)
- **Containerizacao:** Docker / docker-compose

---

## Logging
- Logger baseado em `log/slog` (ver `internal/shared/observability`).
- Config por env: `LOG_LEVEL` (`debug|info|warn|error`) e `LOG_FORMAT` (`text|json|pretty`).
- Por request, o middleware injeta um logger no `context.Context` (inclui `request_id`, método, path e rota quando disponível).

## Configuracao de ambiente (dev/prod)

- Para dev local, copie o exemplo: `cp .env.example .env`.
- O app **nao** carrega `.env` automaticamente. Exporte as variaveis no shell (ou use `direnv`).
- A aplicacao usa `API_HOST` para definir o host da API em cada ambiente.

---

Comandos:

```bash
make dev-api
```

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
│   │   │   └── http/               # HTTP adapter (API)
│   │   │       ├── api/            # API JSON
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
└── README.md                       # Este arquivo
```

