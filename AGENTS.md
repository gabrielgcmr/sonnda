<!-- AGENTS.md -->
# AGENTS.md

Simple instructions for coding agents working on this repo.

## General
- The project is being done by a solo developer.
- When suggesting a solution to a problem, offer the correct way (even if it requires refactoring) and a simple way to solve it.
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
  - **Auth**: Firebase
  - **Database**: PostgreSQL (Supabase managed)
  - **Code generation**: sqlc for type-safe SQL
  - **External integrations**: Google Cloud Document AI, Firebase Auth, Google Cloud Storage
- Web:
  - templ + Tailwind CSS v4 + HTMX 
    - templates em `internal/adapters/inbound/http/web/templates/`
    - statics em `internal/adapters/inbound/http/web/public/`
    - **CSS generation**: Tailwind v4 (source: `internal/adapters/inbound/http/web/styles/`, output: `public/css/app.css`)
- Mobile (Not in this repo):
  - React Native + Expo
- **Development tools**:
  - Air (live reload)
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
  - **Error contract (`internal/app/apperr`)**: Centralized `AppError` codes/messages; handlers must convert via HTTP layer helpers.
  - **Config (`internal/app/config`)**: Env config
  - **bootstrap (`internal/app/bootstrap`)**: Wiring of dependencies, env/config loading.
  - **Observability (`internal/app/observability`)**: Logging setup (slog), request-scoped logger injection.
- **Adapters (`internal/adapters`)**: Concrete implementations and protocol adapters (inbound and outbound).
  - **Inbound (`internal/adapters/inbound/http`)**: HTTP protocol adapter.
    - **API (`internal/adapters/inbound/http/api`)**: RESTful routes, handlers, and middleware for API consumers.
    - **Web (`internal/adapters/inbound/http/web`)**: Server-rendered routes, handlers, and middleware for web UI (Templ + HTMX).
  - **Outbound (`internal/adapters/outbound`)**: Concrete implementations (database repositories, external integrations, cloud services).

## Static files and templates (MANDATORY)
- **Do NOT create** a top-level `assets/` directory in this repository.
- The project intentionally separates responsibilities:
  - `templates/` → server-side rendered UI (`.templ`, layout, components, features)
  - `public/` → static files served directly to the browser (CSS, JS, images, fonts)
- Static files are exposed via the public URL prefix:
  - `/static/*` → mapped to `public/*`
- **Do NOT rewrite paths** in templates or scripts to `/assets/...`.
  - If an automated suggestion mentions `/assets`, replace it with `/static`.
- This decision is documented in:
  **ADR-007 — Separação entre templates e arquivos estáticos (sem pasta /assets)**.

## Tailwind CSS (WEB) - v4 best practices
- Use Tailwind CSS **v4 utilities-first** in `.templ` templates; only add custom CSS when Tailwind cannot express it cleanly.
- **Do not edit generated CSS**: `internal/adapters/inbound/http/web/public/css/app.css` is generated.
- Edit Tailwind source files instead:
  - `internal/adapters/inbound/http/web/styles` (Tailwind entrypoint, `@import "tailwindcss";`, `@theme` tokens)
- Prefer composing UI via reusable `templ` components over writing large custom CSS blocks.
- Avoid `@apply` except for small, reusable abstractions that reduce repeated class strings.
- Keep naming consistent and semantic: use the token system (e.g. colors from `@theme`) instead of hardcoded hex/rgb in templates.
- Build/watch commands:
  - `make dev-web` (Air + Tailwind watch + templ watch) - Preference
  - `make tailwind` (build)
  - `make tailwind-watch` (watch)

---

## Error Handling (MANDATORY)

This project uses a **centralized error contract** based on `AppError`.

### Core rules
- **Do NOT return raw strings as error contracts.**
- **Do NOT expose `err.Error()` in HTTP responses.**
- **Do NOT manually build error JSON in handlers or middleware.**

### AppError
- Application-level errors must be represented as `*apperr.AppError`.
- `AppError` contains:
  - `Code` (`ErrorCode`) - stable, machine-readable contract
  - `Message` - safe, human-readable message
  - `Cause` - optional internal error (wrapped with `%w`)
- Services/use cases **must return `AppError` for known failures** (validation, conflicts, not found, infra errors).
- Domain **never** imports HTTP, Gin, or `apperr`.
- Handlers and middlewares **must call**: httperrors.WriteError(c,err)
- Location: `internal/app/apperr/error.go`
- HTTP error presentation id centralized in:`internal/http/errors/error.go`.
 
 ---

## Logging
- The app uses `log/slog` via `internal/app/config/observability` (request-scoped logger is injected by HTTP middleware).
- Configure with `LOG_LEVEL` (`debug|info|warn|error`) and `LOG_FORMAT` (`text|json|pretty`).

### Access Log
- Exactly one access log entry per request.
- Includes:
  - request_id
  - method, path/route
  - status
  - latency
  - error_code (if present)

### Error Log Policy
- 4xx errors:
  - No detailed error logging in handlers/middleware.
  - AccessLog entry only.
- 5xx errors:
  - AccessLog entry
  - One detailed log with full error chain (Cause) emitted by the HTTP error writer.
- panic
  - Handled by Recovery middleware with stacktrace.

## Test
- Add or update tests when behavior changes.

