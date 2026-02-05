<!-- docs/api/README.md -->
# API Sonnda

Documentação das APIs HTTP da Sonnda (REST).

## Visão geral

- **Base API**: `https://api.sonnda.com.br/v1`
- **Formato de erro**: contrato `AppError` (sempre em `{ "error": { "code", "message" } }`)
- **JSON**: campos em `snake_case`

## Contrato oficial

- A fonte de verdade é `internal/api/openapi/openapi.yaml` (embutido no binário e servido em `/openapi.yaml`).
- A UI em `/docs` é gerada a partir do OpenAPI (Redoc via CDN).
- Arquivos `.md` são guias com exemplos e fluxos, e não duplicam o contrato.
- Para validar o contrato localmente: `make openapi-validate`.

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
