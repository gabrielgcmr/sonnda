<!-- docs/api/errors.md -->
# Erros (RFC 9457 - Problem Details)

A API retorna erros no formato **Problem Details** (RFC 9457) com `Content-Type: application/problem+json`.

Além dos campos padrão (`type`, `title`, `status`, `detail`, `instance`), a Sonnda inclui:

- `code`: código estável (`apperr.ErrorCode`)
- `violations`: lista de violações (quando aplicável)
- `traceId`: id de rastreamento (normalmente o `X-Request-ID`)
- `timestamp`: timestamp (UTC)

## Exemplo (401)

```json
{
  "type": "urn:sonnda:problem:auth_required",
  "title": "Não autorizado",
  "status": 401,
  "detail": "autenticação necessária",
  "instance": "urn:sonnda:request-id:8a0f8a9b-2e1c-4c46-a2b1-1a6f8a6c2e44",
  "code": "AUTH_REQUIRED",
  "traceId": "8a0f8a9b-2e1c-4c46-a2b1-1a6f8a6c2e44",
  "timestamp": "2026-02-06T12:34:56Z"
}
```

## Exemplo (400 com violations)

```json
{
  "type": "urn:sonnda:problem:validation_failed",
  "title": "Falha de validação",
  "status": 400,
  "detail": "entrada inválida",
  "code": "VALIDATION_FAILED",
  "violations": [
    { "field": "birth_date", "reason": "required" }
  ]
}
```

## Perfil da aplicação ausente

`GET /v1/me` e as rotas que exigem cadastro retornam HTTP `403` com
`code: "PROFILE_NOT_FOUND"` quando o token é válido, mas não existe perfil local.
Os clientes devem manter a sessão e abrir o onboarding somente para esse código.
Um `404` genérico, `ACCESS_DENIED`, falha de rede ou erro de servidor não significa
perfil ausente e deve permitir tentar novamente, sem logout automático.

`POST /v1/me` exige apenas autenticação e cria o perfil. Depois de receber `201`,
o cliente pode liberar a aplicação. O Supabase Auth é responsável pela sessão;
a API é a fonte do perfil e das validações do cadastro.
