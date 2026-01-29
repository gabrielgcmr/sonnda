<!-- docs/api/README.md -->
# API Sonnda

Documentação das APIs HTTP da Sonnda (REST).

## Visão geral

- **Base API**: `/v1`
- **Formato de erro**: contrato `AppError` (sempre em `{ "error": { "code", "message" } }`)
- **JSON**: campos em `snake_case`

## Autenticação (resumo)

- **Apps (mobile/backoffice/SPA)**: `Authorization: Bearer <access_token>` em todas as rotas `/v1`.

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
