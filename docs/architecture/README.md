<!-- docs/architecture/README.md -->
# Architecture

Descrição da arquitetura da Sonnda API em camadas, fluxos principais e decisões relevantes.

Este documento descreve **como a arquitetura está organizada**.  
As decisões não óbvias (o *porquê*) são registradas separadamente em ADRs.

---

## Visão geral

O backend segue um modelo em camadas simples, com baixo acoplamento e separação clara de responsabilidades, aplicado de forma pragmática em Go.

- **Domain (`internal/domain`)**  
  Modelos do domínio, regras de negócio e invariantes.  
  - Models em `internal/domain/model`; ports em `internal/domain/ports`.

- **App (`internal/app`)**  
  Orquestração e cross-cutting concerns.  
  - Use cases em `internal/app/usecase`; services em `internal/app/services`.  
  - Erros centralizados em `internal/shared/apperr`.  
  - Config/bootstrapping em `internal/app/config` e `internal/app/bootstrap`.  
  - Observabilidade em `internal/shared/observability` (slog, logger por request).

- **Adapters (`internal/adapters`)**  
  Implementações concretas e adapters de protocolo.  
  - **Inbound (`internal/adapters/inbound/http`)**: HTTP (API REST + Web Templ/HTMX), rotas, handlers e middlewares.  
  - **Outbound (`internal/adapters/outbound`)**: persistência (sqlc/pgx), integrações externas (Firebase Auth, GCS, Document AI).

Essas camadas representam **limites conceituais**, não apenas organização de pastas.

---

## Fluxo de request

1) **Middleware** autentica o usuário e adiciona informações ao contexto  
   (request_id, usuário autenticado, etc.).

2) **Handler HTTP**  
   - valida payload  
   - faz parsing de parâmetros  
   - monta o input do service  

3) **Service / Use case (camada App)**  
  - executa regras de negócio  
  - aplica políticas de acesso  
  - coordena chamadas a repositórios e serviços externos  

4) **Repository (Outbound)**  
  - executa queries via sqlc/pgx  
  - persiste ou consulta dados  

5) **Resposta HTTP**  
   - erros são normalizdos para um contrato estável via `internal/shared/apperr`
      - Veja `docs/architecture/error-handling.md`

---

## Persistência

- SQL definido em `internal/adapters/outbound/storage/data/postgres/sqlc/sql`.
- `sqlc` gera código em `internal/adapters/outbound/storage/data/postgres/sqlc/generated`.
- Repositórios em `internal/adapters/outbound/storage/data/postgres/repository` encapsulam o acesso ao banco.
- Banco principal: PostgreSQL (Supabase).
- Soft delete usa `deleted_at`; consultas filtram `deleted_at IS NULL`.

---

## Observabilidade

- Logger baseado em `log/slog` (`internal/shared/observability`).
- Variáveis:
  - `LOG_LEVEL`
  - `LOG_FORMAT`
- Um logger por request é injetado via middleware HTTP.

---

## Configuração

- Variáveis de ambiente definidas no ambiente (veja `.env.example` para referência).
- `APP_ENV` define o ambiente (`dev | prod`).
- Configurações carregadas na inicialização da aplicação (`internal/app/config`).

---

## Bootstrap e rotas

- Bootstrap faz o wiring (repos, services e handlers) em `internal/app/bootstrap`.
- As rotas HTTP vivem em `internal/adapters/inbound/http/router.go` (API + Web).  
- Níveis de acesso:
  - público
  - autenticado
  - registrado

---

## Decisões arquiteturais (ADR)

Algumas decisões importantes do projeto **não são óbvias apenas pela leitura do código**.  
Para preservar o contexto dessas escolhas ao longo do tempo, o Sonnda adota o uso de **Architecture Decision Records (ADR)**.

Os ADRs documentam:
- o contexto da decisão
- a decisão tomada
- alternativas consideradas
- consequências

Os ADRs vivem em:
`docs/architecture/adr/`.

---

## Controle de acesso (ReBAC)

O projeto segue a direcao de **ReBAC** (Relationship-Based Access Control) para decidir acesso a recursos de paciente.

Em paralelo, existe **RBAC por acao** para limitar o que um tipo de conta pode fazer (AccountType + Professional.Kind).

Detalhes: `docs/architecture/access-control.md`.

Modelo de conta: `docs/architecture/account-model.md`.
