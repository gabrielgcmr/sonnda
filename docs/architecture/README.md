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
  - Entities em `internal/domain/entity`; repositories em `internal/domain/repository`; storage abstractions em `internal/domain/storage`; AI abstractions em `internal/domain/ai`.

- **Application (`internal/application`)**  
  Orquestração e cross-cutting concerns.  
  - Use cases em `internal/application/usecase`; services em `internal/application/services`.  
  - Bootstrapping (injeção de dependências) em `internal/application/bootstrap`.

- **API (`internal/api`)**  
  Implementações de adapters HTTP (inbound).  
  - Handlers em `internal/api/handlers`; rotas em `internal/api/routes`; middlewares em `internal/api/middleware`; presenter em `internal/api/presenter`.

- **Infrastructure (`internal/infrastructure`)**  
  Implementações concretas de persistência e integrações externas.  
  - **Persistence (`internal/infrastructure/persistence`)**: repositórios (sqlc/pgx), cache.
  - **Auth (`internal/infrastructure/auth`)**: Supabase auth provider.
  - **AI (`internal/infrastructure/ai`)**: Google Cloud Document AI adapter.

- **Kernel (`internal/kernel`)**  
  Preocupações transversais (cross-cutting concerns).  
  - Error contract (`internal/kernel/apperr`): `AppError` e catalog de códigos.
  - Observability (`internal/kernel/observability`): logging (slog) com escopo de requisição.
  - **Auth (`internal/infrastructure/auth`)**: autenticação e autorização (Supabase Auth, etc).

- **Kernel (`internal/kernel`)**  
  Núcleo transversal do sistema.
  - **Error contract (`internal/kernel/apperr`)**: contrato centralizado de erros.
  - **Observability (`internal/kernel/observability`)**: logging baseado em slog, logger por request.

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
   - erros são normalizados para um contrato estável via `internal/kernel/apperr`
      - Veja `docs/architecture/error-handling.md`

---

## Persistência

- SQL definido em `internal/infrastructure/persistence/postgres/sqlc/sql`.
- `sqlc` gera código em `internal/infrastructure/persistence/postgres/sqlc/generated`.
- Repositórios em `internal/infrastructure/persistence/postgres/repo` encapsulam o acesso ao banco.
- Banco principal: PostgreSQL (Supabase).
- Soft delete usa `deleted_at`; consultas filtram `deleted_at IS NULL`.

---

## Observabilidade

- Logger baseado em `log/slog` (`internal/kernel/observability`).
- Variáveis:
  - `LOG_LEVEL`
  - `LOG_FORMAT`
- Um logger por request é injetado via middleware HTTP.

---

## Configuração

- Variáveis de ambiente definidas no ambiente (veja `.env.example` para referência).
- `APP_ENV` define o ambiente (`dev | prod`).
- Configurações carregadas na inicialização da aplicação (`internal/config`).

---

## Bootstrap e rotas

- Bootstrap faz o wiring (repos, services e handlers) em `internal/application/bootstrap`.
- As rotas HTTP vivem em `internal/api/routes.go` (API REST).  
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
