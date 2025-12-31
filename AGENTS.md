# AGENTS.md

Simple instructions for coding agents working on this repo.

## General
- Keep changes minimal and focused on the request.
- Prefer small, readable functions and clear naming.
- Avoid editing generated files unless explicitly asked.
- Do not touch secrets or files under `secrets/`.
- Use the logging already configured in the application.

## Logging
- The app uses `log/slog` via `internal/app/config/observability` (request-scoped logger is injected by HTTP middleware).
- Configure with `LOG_LEVEL` (`debug|info|warn|error`) and `LOG_FORMAT` (`text|json|pretty`).
- Prefer `applog.FromContext(ctx)` inside use cases/repos (it carries `request_id`, route/method/path, etc.).
- For API errors that should appear in access logs, set `c.Set("error_code", "<code>")` on the `gin.Context`.

## Go
- Run `gofmt` on Go files you change.
- Keep error handling explicit and consistent with nearby code.
- Add or update tests when behavior changes.

## Reviews and checks
- Call out any assumptions or open questions before finishing.
