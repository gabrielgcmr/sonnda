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
   - Example: // internal/features/patient/profile/service.go.
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
- The architecture is migrating incrementally from global layers to business contexts under `internal/features`, preserving separation of concerns.
- **Features (`internal/features`)**: Context-specific application flows.
  - **Account (`internal/features/account`)**: Profile services, onboarding, DTOs and error mapping; HTTP handler and middleware live in `account/http`.
  - **Patient profile (`internal/features/patient/profile`)**: Patient registration services, DTOs, error mapping and repository contract; its domain model lives in `profile/domain`, HTTP handler in `profile/http`, and Postgres adapter in `profile/postgres`.
  - **Patient access (`internal/features/patient/access`)**: Account-to-patient grants, accessible-patient listing, relationship metadata, application and HTTP services, repository contracts and persistence. Access determines whether an account is linked to a patient; it does not define action-level authorization.
  - Account owns its repository interface and user persistence errors in `account/repository.go`; its Postgres adapter lives in `account/postgres`.
  - User entities and account types belong to `internal/features/account/domain` (package `accountdomain`). Other contexts may import this pure domain package without depending on account application services.
- The shared database client and generated sqlc code remain in `internal/infrastructure` during this migration.
- OpenAPI sources belong to the sibling `../sonnda-contracts`; API generators consume `../sonnda-contracts/dist/openapi.yaml`. Generated Go types live in `internal/generated/openapi`, and the embedded HTTP spec lives in `internal/openapispec`.
  - The shared persistence failure sentinel belongs to `internal/domain/repository/errors.go`; account application code must not import concrete Postgres repositories.
  - `internal/application/bootstrap/account.go` composes the account handler and middleware in a single `AccountModule`; patient access is composed independently in `PatientAccessModule`.
  - Add account behavior to this feature, not to the former global user service, registration use case or user handler paths.
  - Other contexts keep their existing organization until explicitly migrated.
- **Domain (`internal/domain`)**: Core business models and rules (infrastructure and HTTP agnostic).
  - **Entity (`internal/domain/entity`)**: Core business entities.
  - **Repository (`internal/domain/repository`)**: Domain repository interfaces.
  - **Storage (`internal/domain/storage`)**: Storage interfaces (file storage abstractions).
  - **Lab Extraction (`internal/domain/labextraction`)**: contract for structured lab report extraction.
- **Application (`internal/application`)**: Where orchestration and cross-cutting concerns live.
  - **Use cases (`internal/application/usecase`)**: Business flows composed from domain models/ports.
  - Patient creation is coordinated by `internal/application/usecase/patientcreation`: profile data and the creator's explicit relationship are validated by their owning features and persisted atomically.
  - **Services (`internal/application/services`)**: Application services that coordinate repositories/integrations.
  - Patient access checks live in `internal/features/patient/access`: `Checker` permits the patient owner or an account with an active grant. No action, account-type or professional-kind policies are currently implemented.
  - **Bootstrap (`internal/application/bootstrap`)**: Wiring of dependencies, env/config loading.
- **Config (`internal/config`)**: Environment configuration.
- **API (`internal/api`)**: HTTP layer (RESTful API).
  - **Handlers (`internal/api/handlers`)**: HTTP request handlers.
  - **Middleware (`internal/api/middleware`)**: HTTP middlewares (auth, logging, CORS, etc).
  - **Helpers (`internal/api/helpers`)**: HTTP helper functions (binding, validation, identity).
  - **Presenter (`internal/api/presenter`)**: Response formatting and error presentation.
- **Infrastructure (`internal/infrastructure`)**: Concrete implementations and outbound integrations.
  - **Persistence (`internal/infrastructure/persistence`)**: Database repositories, cache, file storage.
  - **Auth (`internal/infrastructure/auth`)**: Authentication provider implementations.
  - **Document AI (`internal/infrastructure/documentai`)**: current Google Cloud Document AI integration.
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
- Feature HTTP handlers and middleware must use `presenter.ErrorResponder(c, err)` and retain the shared access/error logging policy.

---

## Logging
- The app uses `log/slog` via `internal/kernel/observability` (request-scoped logger is injected by HTTP middleware).
- Configure with `LOG_LEVEL` (`debug|info|warn|error`) and `LOG_FORMAT` (`text|json|pretty`).

