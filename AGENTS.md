# AGENTS.md

Simple instructions for coding agents working on this repo.

## General
- The project is being done by a solo developer who is learning programming.
- When suggesting a solution to a problem, offer the correct way (even if it requires refactoring) and a simple way to solve it.
- Prefer small, readable functions and clear naming.
- Avoid editing generated files unless explicitly asked.
- Do not touch secrets or files under `secrets/`.
- Follow the existing error-handling and logging architecture described below.

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
  - `Code` (`ErrorCode`) — stable, machine-readable contract
  - `Message` — safe, human-readable message
  - `Cause` — optional internal error (wrapped with `%w`)
- Services/use cases **must return `AppError` for known failures** (validation, conflicts, not found, infra errors).
- Domain **never** imports HTTP, Gin, or `apperr`.
- Handlers and middlewares **must call**: httperrors.WriteError(c,err)
- Location: `internal/app/apperr/error.go`
- HTTP error presentation id centralized in:`internal/http/errors/error.go`.


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

## Go
- Run `gofmt` on Go files you change.
- Keep error handling explicit and consistent with nearby code.
- Add or update tests when behavior changes.
- 

## Reviews and checks
- Call out any assumptions or open questions before finishing.
- Verify services return AppError for all known failure paths.
