<!-- internal/adapters/inbound/http/web/AGENTS.md -->
# AGENTS.md

Simple instructions for coding in web package

## Static files and templates (MANDATORY)
- Do NOT create a top-level `assets/` directory in this repository.
- The project intentionally separates responsibilities:
  - `templates/` -> server-side rendered UI (`.templ`, layout, components, features)
  - `public/` -> static files served directly to the browser (CSS, JS, images, fonts)
- Static files are exposed via the public URL prefix:
  - `/static/*` -> mapped to `public/*`
- Do NOT rewrite paths in templates or scripts to `/assets/...`.
  - If an automated suggestion mentions `/assets`, replace it with `/static`.
- This decision is documented in: ADR-007 - Separacao entre templates e arquivos estaticos (sem pasta /assets).

## Error handling (MANDATORY)
- `internal/adapters/inbound/http/shared/httperr` is the shared HTTP error presenter (it is NOT API-only).
- WEB has two types of endpoints:
  - SSR pages: return HTML (templ). Prefer redirect on auth/registration failures (no JSON).
  - XHR endpoints (fetch/HTMX): return JSON/204 and use `httperr.WriteError(c, err)` for errors.
- If an endpoint supports both navigation and XHR, use content negotiation (e.g., `Accept: text/html`) to redirect for navigation and return 204/JSON for XHR.

## Tailwind CSS (WEB) - v4 best practices
- Use Tailwind CSS v4 utilities-first in `.templ` templates; only add custom CSS when Tailwind cannot express it cleanly.
- Do not edit generated CSS: `internal/adapters/inbound/http/web/public/css/app.css` is generated.
- Edit Tailwind source files instead:
  - `internal/adapters/inbound/http/web/styles` (Tailwind entrypoint, `@import "tailwindcss";`, `@theme` tokens)
- Prefer composing UI via reusable `templ` components over writing large custom CSS blocks.
- Keep naming consistent and semantic: use the token system (e.g. colors from `@theme`) instead of hardcoded hex/rgb in templates.
- Build/watch commands:
  - `make dev-web` (Air + Tailwind watch + templ watch) - Preference
  - `make tailwind` (build)
  - `make tailwind-watch` (watch)

---