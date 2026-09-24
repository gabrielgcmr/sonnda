<!-- internal/features/account/AGENTS.md -->
# Account

- Keep profile services, onboarding and application DTOs in this package. It must not import Gin or the HTTP layer.
- Keep HTTP binding, request identity and response presentation in `http/`.
- Return known failures through `internal/kernel/apperr`; HTTP handlers and middleware call `presenter.ErrorResponder(c, err)`.
- Preserve exactly one access log per request. Do not log 4xx details in handlers; the shared presenter writes the detailed 5xx error log.
- Compose dependencies in `internal/application/bootstrap/account.go`.
- This is migration stage 2 for persistence: the repository interface and user persistence errors live in `repository.go`; the concrete SQLC/pgx adapter lives in `postgres/`.
- Application code depends on `Repository`, never on the Postgres adapter or the legacy concrete repositories. Generic persistence failures use `internal/domain/repository.ErrRepositoryFailure`.
- User entities, patient access interfaces, the shared database client and generated sqlc code stay in their existing packages.
- The professional service and patient access dependencies remain transitional dependencies for a later migration.
- Keep `/v1/me`, the `/v1/users` creation alias and `/v1/me/patients` behavior compatible during structural moves.
