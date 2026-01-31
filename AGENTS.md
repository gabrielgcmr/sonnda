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
   - Example: // internal/application/services/patient/service.go. 
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
  - **Entity (`internal/domain/entity`)**: Core business entities.
  - **Repository (`internal/domain/repository`)**: Domain repository interfaces.
  - **Storage (`internal/domain/storage`)**: Storage interfaces (file storage abstractions).
  - **AI (`internal/domain/ai`)**: AI/ML interfaces (abstractions for document processing).
- **Application (`internal/application`)**: Where orchestration and cross-cutting concerns live.
  - **Use cases (`internal/application/usecase`)**: Business flows composed from domain models/ports.
  - **Services (`internal/application/services`)**: Application services that coordinate repositories/integrations.
  - **Bootstrap (`internal/application/bootstrap`)**: Wiring of dependencies, env/config loading.
- **Config (`internal/config`)**: Environment configuration.
- **API (`internal/api`)**: HTTP layer (RESTful API).
  - **Handlers (`internal/api/handlers`)**: HTTP request handlers.
  - **Middleware (`internal/api/middleware`)**: HTTP middlewares (auth, logging, CORS, etc).
  - **Helpers (`internal/api/helpers`)**: HTTP helper functions (binding, validation, identity).
  - **Presenter (`internal/api/presenter`)**: Response formatting and error presentation.
- **Infrastructure (`internal/infrastructure`)**: Concrete implementations and outbound integrations.
  - **Persistence (`internal/infrastructure/persistence`)**: Database repositories, cache, file storage.
  - **Auth (`internal/infrastructure/auth`)**: Authentication/authorization implementations.
  - **AI (`internal/infrastructure/ai`)**: AI integrations (Google Cloud Document AI).
- **Kernel (`internal/kernel`)**: Cross-cutting concerns.
  - **Error contract (`internal/kernel/apperr`)**: Centralized `AppError` codes/messages; handlers must convert via HTTP layer helpers.
  - **Observability (`internal/kernel/observability`)**: Logging setup (slog), request-scoped logger injection.

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
- Services/use cases **must return `AppError` for known failures** (validation, conflicts, not found, infra errors).
- Domain **never** imports HTTP, Gin, or `apperr`.
- Handlers and middlewares **must call**: `presenter.WriteError(c, err)` or similar helper from the presenter layer.
- HTTP error presentation is centralized in: `internal/api/presenter`.

---

## Logging
- The app uses `log/slog` via `internal/kernel/observability` (request-scoped logger is injected by HTTP middleware).
- Configure with `LOG_LEVEL` (`debug|info|warn|error`) and `LOG_FORMAT` (`text|json|pretty`).

## Test
- Add or update tests when behavior changes.
