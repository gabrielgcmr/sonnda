# Project Context

## Stack
- Backend: Go
- Domain: internal/core/domain
- DB: sqlc + Supabase
- Frontend:
- Android: Kotlin (implementing)
- Ios: Swift (not started)
- Web: HTMX + http/template + Tailwind CSS

## Recent Decisions
- Domain structs have no tags.
- IDs are strings (no UUID types in domain).
- Exams are separated by type under medicalRecord/exam.
- Official lab record type is LABS_EXAM.

## Domain Layout (current)
- identity: users, authorizations, errors
- patient: patient entity + errors
- demographics: CPF, gender, race
- medicalRecord: MedicalRecord aggregate
- medicalRecord/prevention, medicalRecord/problem
- medicalRecord/exam/lab
- medicalRecord/exam/physical
- medicalRecord/exam/image (placeholder)

## Usecases (current)
- lab usecases moved to internal/core/usecases/lab (package lab)

## sqlc Changes (done)
- app_users and patients tables no longer default to gen_random_uuid().
- INSERTs for app_users, patients, lab_reports, lab_results, lab_result_items now require explicit id.

## Web (HTMX + Tailwind) Setup
- Tailwind config: tailwind.config.js (content points to web templates + handlers)
- Tailwind input: internal/adapters/inbound/http/web/static/css/input.css
- Tailwind output: internal/adapters/inbound/http/web/static/css/tailwind.css
- Standalone Tailwind binary expected at tools/tailwindcss.exe (ignored in git)
- Makefile targets: tailwind, tailwind-watch, dev-web (air + tailwind watch)
- Web routes: internal/adapters/inbound/http/web/routes.go
  - SetupRoutes mounts /static and loads templates
  - parseGlobIfExists prevents panic when a template folder is empty
- Templates:
  - internal/adapters/inbound/http/web/templates/pages/home.html (HTMX hello button)
  - internal/adapters/inbound/http/web/templates/partials/hello.html (fragment)
- main.go registers web routes before API routes

## TODO / Next Steps
- Run `sqlc generate` to refresh generated code after schema/query changes.
- Update supabase repositories to pass string IDs and match new SQL signatures.
- Update remaining ports/usecases/handlers for string IDs (remove UUID parsing).
- Resolve name conflict between usecase package `lab` and domain package `lab` with import aliases.
- Add base layout/templates if needed.
