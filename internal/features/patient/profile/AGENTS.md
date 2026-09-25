<!-- internal/features/patient/profile/AGENTS.md -->
# Patient profile

- Keep patient registration flows, application DTOs, and error mapping in this package.
- Keep HTTP binding, request parsing, and response presentation in `http/`.
- Keep the patient repository interface and patient-specific persistence errors in `repository.go`.
- Keep the concrete pgx/sqlc adapter in `postgres/`; shared database clients and generated sqlc code remain in `internal/infrastructure`.
- Preserve the existing patient endpoints and response contracts while this feature is migrated.
- Patient access remains a transitional dependency until the `patient/access` feature is migrated.
- Patient domain entities remain in `internal/domain/entity/patient` during migration stage 2.
- Return known failures through `internal/kernel/apperr`; HTTP handlers call `presenter.ErrorResponder(c, err)`.
- Compose dependencies in `internal/application/bootstrap/patient.go`.
