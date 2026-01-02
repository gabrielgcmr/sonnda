# Architecture

Descrição da arquitetura da Sonnda API em camadas, fluxos principais e decisões relevantes.

Este documento descreve **como a arquitetura está organizada**.  
As decisões não óbvias (o *porquê*) são registradas separadamente em ADRs.

---

## Visão geral

O backend segue um modelo em camadas simples, com baixo acoplamento e separação clara de responsabilidades, inspirado em Clean Architecture, mas aplicado de forma pragmática e idiomática em Go.

- **Domain (`internal/domain`)**  
  Entidades, regras de negócio, invariantes e portas (interfaces).

- **App (`internal/app`)**  
  Services de aplicação, políticas de acesso e orquestração de fluxos.  
  Não há use cases individuais; os **Services representam o boundary da aplicação**.

- **HTTP (`internal/http`)**  
  Rotas, handlers e middlewares.  
  Responsável por autenticação, validação de payloads e exposição controlada de dados.

- **Infrastructure (`internal/infrastructure`)**  
  Integrações externas (auth, persistence, storage, document AI, etc.).

Essas camadas representam **limites conceituais**, não apenas organização de pastas.

---

## Fluxo de request

1) **Middleware** autentica o usuário e adiciona informações ao contexto  
   (request_id, usuário autenticado, etc.).

2) **Handler HTTP**  
   - valida payload  
   - faz parsing de parâmetros  
   - monta o input do service  

3) **Service (camada App)**  
   - executa regras de negócio  
   - aplica políticas de acesso  
   - coordena chamadas a repositórios e serviços externos  

4) **Repository (Infrastructure)**  
   - executa queries via sqlc/pgx  
   - persiste ou consulta dados  

5) **Resposta HTTP**  
   - erros e dados são mapeados explicitamente no handler  

---

## Persistência

- SQL definido em `internal/infrastructure/persistence/sqlc/sql`.
- O `sqlc` gera código em `internal/infrastructure/persistence/sqlc/generated`.
- Repositórios em `internal/infrastructure/persistence/supabase` encapsulam o acesso ao banco.
- O banco principal é PostgreSQL (Supabase).
- Soft delete utiliza o campo `deleted_at`; consultas filtram `deleted_at IS NULL`.

---

## Observabilidade

- Logger baseado em `log/slog` (`internal/app/config/observability`).
- Variáveis:
  - `LOG_LEVEL`
  - `LOG_FORMAT`
- Um logger por request é injetado via middleware HTTP.

---

## Configuração

- Variáveis de ambiente definidas em `.env` (ver `.env.example`).
- `APP_ENV` define o ambiente (`dev | prod`).
- Configurações são carregadas na inicialização da aplicação.

---

## Bootstrap e rotas

- O bootstrap faz o wiring (repos, services e handlers) em `internal/app/bootstrap`.
- As rotas são definidas em `internal/http/api/router.go`.
- Os níveis de acesso incluem:
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

O projeto segue a direcao de **ReBAC** (Relationship-Based Access Control): acesso e decidido pelo vinculo usuario <-> paciente (membership), e nao por "RBAC + ABAC".

Detalhes: `docs/architecture/access-control.md`.
