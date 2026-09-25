<!-- internal/features/patient/profile/AGENTS.md -->
# Patient profile

- Keep patient registration flows, application DTOs, and error mapping in this package.
- Keep HTTP binding, request parsing, and response presentation in `http/`.
- Keep the patient repository interface and patient-specific persistence errors in `repository.go`.
- Keep the concrete pgx/sqlc adapter in `postgres/`; shared database clients and generated sqlc code remain in `internal/infrastructure`.
- Keep patient entities, validation rules, domain errors, and their tests in `domain/` (package `profiledomain`).
- The domain package must not import profile application services, HTTP, Gin, `apperr`, or infrastructure.
- Preserve the existing patient endpoints and response contracts while this feature is migrated.
- The initial patient grant uses contracts from `patient/access`; the shared access checker remains transitional until migration stage 4.2.
- Return known failures through `internal/kernel/apperr`; HTTP handlers call `presenter.ErrorResponder(c, err)`.
- Compose dependencies in `internal/application/bootstrap/patient.go`.
