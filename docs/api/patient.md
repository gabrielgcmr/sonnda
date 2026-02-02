<!-- docs/api/patient.md -->
# Pacientes

Endpoints para criação e consulta de pacientes.

## Base URL

`https://api.sonnda.com.br/v1`

## Autenticação

Todas as rotas exigem `Authorization: Bearer <id_token>` e usuário registrado.

## Campos e enums

### Create (obrigatórios)

- `cpf` (string) — 11 dígitos
- `full_name` (string)
- `birth_date` (string) — formato `YYYY-MM-DD`
- `gender` — `MALE | FEMALE | OTHER | UNKNOWN`
- `race` — `WHITE | BLACK | ASIAN | MIXED | INDIGENOUS | UNKNOWN`

### Create (opcionais)

- `phone` (string)
- `avatar_url` (string)

### Update (opcionais)

Mesmo conjunto de campos, todos opcionais.

## Endpoints

| Método | Rota            | Status                              |
| ------ | --------------- | ----------------------------------- |
| POST   | `/patients`     | Ativo                               |
| GET    | `/patients`     | Ativo                               |
| GET    | `/patients/:id` | Ativo                               |
| PUT    | `/patients/:id` | Handler existe, rota não registrada |
| DELETE | `/patients/:id` | Handler existe, rota não registrada |

## Criar paciente (POST /v1/patients)

**Request (JSON):**
```json
{
  "cpf": "12345678901",
  "full_name": "Joana Silva",
  "birth_date": "1990-05-12",
  "gender": "FEMALE",
  "race": "WHITE",
  "phone": "+55 11 99999-0000",
  "avatar_url": "https://example.com/avatar.png"
}
```

**Exemplo (curl):**
```bash
curl -i -X POST https://api.sonnda.com.br/v1/patients \
  -H "Authorization: Bearer <id_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "cpf": "12345678901",
    "full_name": "Joana Silva",
    "birth_date": "1990-05-12",
    "gender": "FEMALE",
    "race": "WHITE",
    "phone": "+55 11 99999-0000",
    "avatar_url": "https://example.com/avatar.png"
  }'
```

**Resposta (201 Created):**
```json
{
  "id": "018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11"
}
```

**Erros comuns:**
- `VALIDATION_FAILED` (400)
- `RESOURCE_ALREADY_EXISTS` (409) — CPF duplicado
- `INFRA_DATABASE_ERROR` (5xx)

## Buscar paciente (GET /v1/patients/:id)

**Resposta (200 OK):**
```json
{
  "id": "018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11",
  "owner_user_id": "018f39f2-0b1a-7c5a-9d9e-2b7d8d9c3f11",
  "cpf": "12345678901",
  "cns": null,
  "full_name": "Joana Silva",
  "birth_date": "1990-05-12T00:00:00Z",
  "gender": "FEMALE",
  "race": "WHITE",
  "avatar_url": "https://example.com/avatar.png",
  "phone": "+55 11 99999-0000",
  "created_at": "2026-01-10T12:00:00Z",
  "updated_at": "2026-01-10T12:00:00Z"
}
```

**Exemplo (curl):**
```bash
curl -i https://api.sonnda.com.br/v1/patients/018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11 \
  -H "Authorization: Bearer <id_token>"
```

**Erros comuns:**
- `VALIDATION_FAILED` (400)
- `NOT_FOUND` (404)

## Listar pacientes (GET /v1/patients)

**Resposta (200 OK):**
```json
[
  {
    "id": "018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11",
    "full_name": "Joana Silva",
    "cpf": "12345678901"
  }
]
```

**Exemplo (curl):**
```bash
curl -i https://api.sonnda.com.br/v1/patients \
  -H "Authorization: Bearer <id_token>"
```

**Notas:**
- Paginação está fixa em `limit=100` e `offset=0` no handler atual.
