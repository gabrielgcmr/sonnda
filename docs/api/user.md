<!-- docs/api/user.md -->
# Usuários

Endpoints de perfil e onboarding de usuário.

## Base URL

`https://api.sonnda.com.br/v1`

## Autenticação

Todas as rotas exigem `Authorization: Bearer <id_token>`.

## Contrato oficial

O contrato completo de endpoints, schemas e erros fica no OpenAPI: `/openapi.yaml`.

## Criar usuário (POST /v1/users)

Nota (MVP): a criação/gestão de profissionais ainda não está implementada.

**Request (JSON):**
```json
{
  "full_name": "Joana Silva",
  "birth_date": "1990-05-12",
  "cpf": "12345678901",
  "phone": "+55 11 99999-0000"
}
```

**Exemplo (curl):**
```bash
curl -i -X POST https://api.sonnda.com.br/v1/users \
  -H "Authorization: Bearer <id_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Joana Silva",
    "birth_date": "1990-05-12",
    "cpf": "12345678901",
    "phone": "+55 11 99999-0000"
  }'
```

## Perfil atual (GET /v1/me)

**Resposta (200 OK):**
```json
{
  "id": "018f39f2-0b1a-7c5a-9d9e-2b7d8d9c3f11",
  "auth_provider": "supabase",
  "auth_subject": "uid",
  "email": "user@example.com",
  "full_name": "Joana Silva",
  "account_type": "basic_care",
  "birth_date": "1990-05-12T00:00:00Z",
  "cpf": "12345678901",
  "phone": "+55 11 99999-0000",
  "created_at": "2026-01-10T12:00:00Z",
  "updated_at": "2026-01-10T12:00:00Z"
}
```

**Exemplo (curl):**
```bash
curl -i https://api.sonnda.com.br/v1/me \
  -H "Authorization: Bearer <access_token>"
```

## Atualizar perfil (PUT /v1/me)

**Request (JSON):**
```json
{
  "full_name": "Joana S. Silva",
  "birth_date": "1990-05-12",
  "cpf": "12345678901",
  "phone": "+55 11 99999-0000"
}
```

**Exemplo (curl):**
```bash
curl -i -X PUT https://api.sonnda.com.br/v1/me \
  -H "Authorization: Bearer <id_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Joana S. Silva",
    "birth_date": "1990-05-12",
    "cpf": "12345678901",
    "phone": "+55 11 99999-0000"
  }'
```

## Deletar perfil (DELETE /v1/me)

**Exemplo (curl):**
```bash
curl -i -X DELETE https://api.sonnda.com.br/v1/me \
  -H "Authorization: Bearer <id_token>"
```

## Listar pacientes do usuário (GET /v1/me/patients)

Parâmetros opcionais: `limit`, `offset`.

**Exemplo (curl):**
```bash
curl -i "https://api.sonnda.com.br/v1/me/patients?limit=20&offset=0" \
  -H "Authorization: Bearer <id_token>"
```
