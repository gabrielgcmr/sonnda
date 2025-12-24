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
- sqlc generate and build are stable (resolved).
- Supabase/sqlc repositories adjusted for required IDs (resolved).
- Ports/usecases/handlers updated for string IDs, UUID parsing removed (resolved).
- Lab package name conflict resolved with standardized naming (resolved).

## Web (HTMX + Tailwind) Setup
- Tailwind config: tailwind.config.js (content points to web templates + handlers)
- Tailwind input: internal/adapters/inbound/http/web/static/css/input.css
- Tailwind output: internal/adapters/inbound/http/web/static/css/tailwind.css
- Tailwind v4 setup uses `@import "tailwindcss"` and `@theme` tokens mapped to Material CSS vars.
- Standalone Tailwind binary expected at tools/tailwindcss.exe (ignored in git)
- Makefile targets: tailwind, tailwind-watch, dev-web (air + tailwind watch with config)
- Web routes: internal/adapters/inbound/http/web/routes.go
  - SetupRoutes mounts /static via go:embed and loads templates
- Templates (layout + pages):
  - internal/adapters/inbound/http/web/templates/layouts/base.html (shared layout + theme toggle)
  - internal/adapters/inbound/http/web/templates/layouts/app.html (protected layout with header + user name)
  - internal/adapters/inbound/http/web/templates/pages/home.html
  - internal/adapters/inbound/http/web/templates/pages/login.html
  - internal/adapters/inbound/http/web/templates/pages/signup.html
  - internal/adapters/inbound/http/web/templates/pages/dashboard.html
  - internal/adapters/inbound/http/web/templates/partials/hello.html (fragment)
  - internal/adapters/inbound/http/web/templates/partials/scripts.html (shared scripts include)
- Theme tokens:
  - internal/adapters/inbound/http/web/static/css/theme.css (Material light/dark vars)
  - input.css imports theme.css and declares @theme color tokens
- Shared frontend scripts:
  - internal/adapters/inbound/http/web/static/js/app.js (theme toggle + auto-reload)
- Template rendering uses a `render` helper to execute named templates dynamically.

## TODO / Next Steps
- Consider dev-only static serving (filesystem) to avoid rebuilds with go:embed.
