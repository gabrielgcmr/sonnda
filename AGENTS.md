<!-- AGENTS.md -->
# AGENTS.md

Simple instructions for coding agents working on this repo.

## General
- The project is being done by a solo developer.
- When suggesting a solution to a problem, try offer the correct way (even if it requires refactoring) and a simple way to solve it.
- Prefer small, readable functions and clear naming.
- Avoid editing generated files unless explicitly asked.
- Call out any assumptions or open questions before finishing.
- Do not touch secrets or files under `secrets/`.
- Follow the existing error-handling and logging architecture described below.
- Always start every source file you create or modify with a one-line header comment containing the workspace-relative path to that file, formatted as "path/to/file". 
   - Use the language's comment syntax (Go/TS/JS: //, HTML/Markdown: <!-- -->, CSS: /* */). 
   - Example: // internal/app/services/patient/service.go. 
   - Skip only when the format does not support comments or the file is auto-generated.

## Stack
- Backend: API Restful in Go
  - Gin + sqlc + Supabase
  - **Auth**: Supabase
  - **Persistence**: Database: PostgreSQL (Supabase managed) and Redis (Upstash), File Storage: Supabase.
  - **External integrations**: Google Cloud Document AI,

- **Development tools**:
  - Air (live reload)
  - SQLC (SQL code generation)
  - Make (task automation)
  - Docker + docker-compose (containerization)

## Arquitetura
- The architecture follows a layered approach with low coupling and clear separation of concerns:
- **Domain (`internal/domain`)**: Core business models and rules (infrastructure and HTTP agnostic).
  - **Models (`internal/domain/model`)**: Data structures and domain entities.
  - **Ports (`internal/domain/ports`)**: Domain interfaces (abstractions for integrations and repositories).
- **App (`internal/app`)**: Where orchestration and cross-cutting concerns live.
  - **Use cases (`internal/app/usecase`)**: Business flows composed from domain models/ports.
  - **Services (`internal/app/services`)**: Application services that coordinate repositories/integrations.
  - **Error contract (`internal/shared/apperr`)**: Centralized `AppError` codes/messages; handlers must convert via HTTP layer helpers.
  - **Config (`internal/app/config`)**: Env config
  - **bootstrap (`internal/app/bootstrap`)**: Wiring of dependencies, env/config loading.
  - **Observability (`internal/shared/observability`)**: Logging setup (slog), request-scoped logger injection.
- **Adapters (`internal/adapters`)**: Concrete implementations and protocol adapters (inbound and outbound).
  - **Inbound (`internal/adapters/inbound/http`)**: HTTP protocol adapter.
    - **API (`internal/adapters/inbound/http/api`)**: RESTful routes, handlers, and middleware for API consumers.
    - **Shared (`internal/adapters/inbound/http/shared`)**: Shared packages between api and web
    - **Web (`internal/adapters/inbound/http/web`)**: Server-rendered routes, handlers, and middleware for web UI (Templ + HTMX).
  - **Outbound (`internal/adapters/outbound`)**: Concrete implementations (database repositories, external integrations, cloud services).

## Error Handling (MANDATORY)

This project uses a **centralized error contract** based on `AppError`.

### Core rules
- **Do NOT return raw strings as error contracts.**
- **Do NOT expose `err.Error()` in HTTP responses.**
- **Do NOT manually build error JSON in handlers or middleware.**

### AppError
- Application-level errors must be represented as `*apperr.AppError`.
- Location: `internal/kernel/apperr`
- `AppError` contains:
  - `Code` (`ErrorCode`) - stable, machine-readable contract
  - `Message` - safe, human-readable message
  - `Cause` - optional internal error (wrapped with `%w`)
- **Prefer using helper constructors** from `internal/kernel/apperr/factory.go` instead of manually constructing them.
- Services/use cases **must return `AppError` for known failures** (validation, conflicts, not found, infra errors).- Function signatures **may return `error`** for clarity.
- Domain **never** imports HTTP, Gin, or `apperr`.

 
 ---

## Logging
- The app uses `log/slog` via `internal/shared/observability` (request-scoped logger is injected by HTTP middleware).
- Configure with `LOG_LEVEL` (`debug|info|warn|error`) and `LOG_FORMAT` (`text|json|pretty`).

## Test
- Add or update tests when behavior changes.
