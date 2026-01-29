<!-- docs/api/user.md -->
# Usuários

Endpoints de perfil e onboarding de usuário.

## Base URL

`/v1`

## Autenticação

Todas as rotas exigem `Authorization: Bearer <id_token>`.

## Endpoints

| Método | Rota           | Status |
| ------ | -------------- | ------ |
| POST   | `/register`    | Ativo  |
| GET    | `/me`          | Ativo  |
| PUT    | `/me`          | Ativo  |
| DELETE | `/me`          | Ativo  |
| GET    | `/me/patients` | Ativo  |

## Onboarding (POST /v1/register)

**Request (JSON):**
```json
{
  "full_name": "Joana Silva",
  "birth_date": "1990-05-12",
  "cpf": "12345678901",
  "phone": "+55 11 99999-0000",
  "account_type": "professional",
  "professional": {
    "kind": "physician",
    "registration_number": "CRM 12345",
    "registration_issuer": "CRM",
    "registration_state": "SP"
  }
}
```

**Exemplo (curl):**
```bash
curl -i -X POST http://localhost:8080/v1/register \
  -H "Authorization: Bearer <id_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Joana Silva",
    "birth_date": "1990-05-12",
    "cpf": "12345678901",
    "phone": "+55 11 99999-0000",
    "account_type": "professional",
    "professional": {
      "kind": "physician",
      "registration_number": "CRM 12345",
      "registration_issuer": "CRM",
      "registration_state": "SP"
    }
  }'
```

**Erros comuns:**
- `VALIDATION_FAILED` (400)
- `RESOURCE_ALREADY_EXISTS` (409)

## Perfil atual (GET /v1/me)

**Resposta (200 OK):**
```json
{
  "id": "018f39f2-0b1a-7c5a-9d9e-2b7d8d9c3f11",
  "auth_provider": "supabase",
  "auth_subject": "uid",
  "email": "user@example.com",
  "full_name": "Joana Silva",
  "account_type": "professional",
  "birth_date": "1990-05-12T00:00:00Z",
  "cpf": "12345678901",
  "phone": "+55 11 99999-0000",
  "created_at": "2026-01-10T12:00:00Z",
  "updated_at": "2026-01-10T12:00:00Z"
}
```

**Exemplo (curl):**
```bash
curl -i http://localhost:8080/v1/me \
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
curl -i -X PUT http://localhost:8080/v1/me \
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
curl -i -X DELETE http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <id_token>"
```

## Listar pacientes do usuário (GET /v1/me/patients)

Parâmetros opcionais: `limit`, `offset`.

**Exemplo (curl):**
```bash
curl -i "http://localhost:8080/api/v1/me/patients?limit=20&offset=0" \
  -H "Authorization: Bearer <id_token>"
```
