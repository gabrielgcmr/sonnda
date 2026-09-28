<!-- internal/api/AGENTS.md -->
# AGENTS.md

Simple instructions for coding in api package

## Error handling (MANDATORY)
- Huma routes must return Huma errors (`huma.Error*`, `huma.NewError`, or `huma.WriteErr`) and use the standard Huma Problem Details response.
- Gin routes that have not yet migrated use `internal/api/presenter.ErrorResponder`; do not manually build error JSON.
- Known failures must be returned as `*apperr.AppError` from services/usecases; avoid expose `err.Error()` in responses.

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
