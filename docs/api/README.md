<!-- docs/api/README.md -->
# API Sonnda

Documentação das APIs HTTP da Sonnda (REST + endpoints de autenticação web).

## Visão geral

- **Base API**: `/api/v1`
- **Sessão web (cookie)**: `/auth/*`
- **Formato de erro**: contrato `AppError` (sempre em `{ "error": { "code", "message" } }`)
- **JSON**: campos em `snake_case`

## Autenticação (resumo)

- **Apps (mobile/backoffice)**: `Authorization: Bearer <id_token>` em todas as rotas `/api/v1`.
- **Web (HTMX/SSR)**: `POST /auth/session` com `{ "id_token": "..." }` para criar cookie `__session`.

## Índice

- [Autenticação](auth.md)
- [Pacientes](patient.md)
- [Usuários](user.md)
- [Labs](labs.md)

## Contrato de erro (AppError)

Exemplo de resposta:

```json
{
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "autenticação necessária"
  }
}
```
