<!-- docs/api/auth.md -->
# Autenticação

## Visão geral

A Sonnda usa **Supabase Auth**. O cliente autentica no Supabase, recebe um **Access Token** (JWT) e então:

- **API REST (/v1)**: envia o token no header `Authorization: Bearer <access_token>`.

## Arquivos-chave

| Camada | Arquivo |
|---|---|
| Middleware | `internal/adapters/inbound/http/api/middleware/auth.go` |
| Provider Supabase | `internal/adapters/outbound/auth/supabase_bearer_provider.go` |

## Configuração (env)

- `SUPABASE_PROJECT_URL` (obrigatório)
- `SUPABASE_JWT_ISSUER` (opcional, sobrescreve o issuer derivado do projeto)
- `SUPABASE_JWT_AUDIENCE` (opcional)

## API REST (Bearer)

Todas as rotas em `/v1` exigem:

```
Authorization: Bearer <access_token>
```

### Onboarding

Endpoint de cadastro:

```
POST /v1/register
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
