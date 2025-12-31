# Project Context

## Stack
- Backend: API Restful in Go
  - Gin + sqlc + Supabase
  - domain layer + app layer + http layer + infrastructure layer
- Mobile:
  - React Native + Expo
- Web:
  - Health workers: React + Tailwind + Vite

## Recent Decisions
- Protected API routes require registered users with complete profiles.
- Logging uses `log/slog` via `internal/app/config/observability` with `LOG_LEVEL` and `LOG_FORMAT` (`text|json|pretty`); HTTP middleware adds `request_id` correlation.

## Domain Layout (current)
- identity: identity provider/subject/email/claims
- user: user entity + roles + profile completeness rules
- caregiver: caregiver profile (phone) + errors
- professional: professional profile (registration + verification status) + errors
- patient: patient entity + errors
- patientaccess: membership between app user and patient (role + permissions)


## Usecases (current)
- Services live under `internal/app/services` (patient, user, labs).

## API (current)


## TODO / Next Steps
-
