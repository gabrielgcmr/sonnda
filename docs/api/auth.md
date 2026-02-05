<!-- docs/api/auth.md -->
# Autenticação

## Visão geral

A Sonnda usa **Supabase Auth**. O cliente autentica no Supabase, recebe um **Access Token** (JWT) e então:

- **API REST (/v1)**: envia o token no header `Authorization: Bearer <access_token>`.

## Contrato oficial

O contrato de endpoints, schemas e erros fica no OpenAPI: `/openapi.yaml`.

## Arquivos-chave

| Camada | Arquivo |
|---|---|
| Middleware | `internal/api/middleware/auth.go` |
| Provider Supabase | `internal/infrastructure/auth/supabase_bearer_provider.go` |

## Configuração (env)

- `SUPABASE_PROJECT_URL` (obrigatório)
- `SUPABASE_JWT_ISSUER` (opcional, sobrescreve o issuer derivado do projeto)
- `SUPABASE_JWT_AUDIENCE` (opcional)

## Uso do token

Todas as rotas em `/v1` exigem:

```
Authorization: Bearer <access_token>
```

## Fluxo de onboarding

Após autenticar no Supabase, o cliente chama o endpoint de criação de usuário (`POST /v1/users`).

Exemplo (curl):

```bash
curl -i -X POST https://api.sonnda.com.br/v1/users \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```
