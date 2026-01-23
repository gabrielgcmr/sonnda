<!-- docs/api/auth.md -->
# Autenticação

## Visão geral

A Sonnda usa **Firebase Authentication**. O cliente autentica no Firebase, recebe um **ID Token** (JWT) e então:

- **API REST (/api/v1)**: envia o token no header `Authorization: Bearer <id_token>`.
- **Web (SSR/HTMX)**: cria uma sessão HTTP via `POST /auth/session` e recebe o cookie `__session`.

## Arquivos-chave

| Camada | Arquivo |
|---|---|
| Middleware | `internal/adapters/inbound/http/middleware/auth.go` |
| Identidade (port) | `internal/domain/ports/integration/Identity_service.go` |
| Integração Firebase | `internal/adapters/outbound/integrations/auth/firebase_auth_service.go` |

## Sessão web (cookie)

### POST /auth/session

Cria o cookie `__session` a partir do `id_token` do Firebase.

**Request (JSON):**
```json
{
  "id_token": "eyJhbGciOi..."
}
```

**Resposta:**
- `204 No Content` + `Set-Cookie: __session=...`

**Exemplo (curl):**
```bash
curl -i -X POST http://localhost:8080/auth/session \
  -H "Content-Type: application/json" \
  -d '{"id_token":"eyJhbGciOi..."}'
```

**Erros comuns:**
- `VALIDATION_FAILED` (400) — body inválido
- `AUTH_TOKEN_INVALID` (401) — token inválido/expirado

---

### GET /auth/session

**Resposta (200 OK):**
```json
{
  "authenticated": true,
  "user": {
    "provider": "firebase",
    "subject": "uid",
    "email": "user@example.com"
  }
}
```

**Exemplo (curl):**
```bash
curl -i http://localhost:8080/auth/session
```

**Erros comuns:**
- `AUTH_REQUIRED` (401) — cookie ausente
- `AUTH_TOKEN_INVALID` (401) — cookie inválido/expirado

---

### POST /auth/logout

Revoga a sessão e limpa o cookie.

- `204 No Content`

---

### DELETE /auth/session

Alias de logout (idempotente).

- `204 No Content`

---

### POST /auth/session/refresh

Recria a sessão usando novo `id_token` (mesmo payload do `POST /auth/session`).

## API REST (Bearer)

Todas as rotas em `/api/v1` exigem:

```
Authorization: Bearer <id_token>
```

### Onboarding

Endpoint de cadastro:

```
POST /api/v1/register
```

## Firebase (client-side)

Base: `https://identitytoolkit.googleapis.com/v1`

### Criar conta
```
POST /accounts:signUp?key=<FIREBASE_API_KEY>
```

### Login
```
POST /accounts:signInWithPassword?key=<FIREBASE_API_KEY>
```

### Refresh token
```
POST https://securetoken.googleapis.com/v1/token?key=<FIREBASE_API_KEY>
```

## Contrato de erro (AppError)

Exemplo:
```json
{
  "error": {
    "code": "AUTH_TOKEN_INVALID",
    "message": "token inválido ou expirado"
  }
}
```
