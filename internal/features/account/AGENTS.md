<!-- internal/features/account/AGENTS.md -->
# Account

- Keep profile services, onboarding and application DTOs in this package. It must not import Gin or the HTTP layer.
- Keep HTTP binding, request identity and response presentation in `http/`.
- Return known failures through `internal/kernel/apperr`; HTTP handlers and middleware call `presenter.ErrorResponder(c, err)`.
- Preserve exactly one access log per request. Do not log 4xx details in handlers; the shared presenter writes the detailed 5xx error log.
- Compose dependencies in `internal/application/bootstrap/account.go`.
- This is migration stage 1: user entities, repository interfaces, concrete repositories and sqlc stay in their existing packages.
- The existing repository error mapping and professional service dependency remain transitional dependencies. Do not expand persistence responsibilities here.
- Keep `/v1/me`, the `/v1/users` creation alias and `/v1/me/patients` behavior compatible during structural moves.
