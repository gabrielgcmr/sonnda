<!-- docs/api/labs.md -->
# Labs

Endpoints de laudos laboratoriais.

## Base URL

`https://api.sonnda.com.br/v1`

## Autenticação

Todas as rotas exigem `Authorization: Bearer <id_token>`.

## Endpoints

| Método | Rota                        | Status |
| ------ | --------------------------- | ------ |
| GET    | `/patients/:id/labs`        | Ativo  |
| GET    | `/patients/:id/labs/full`   | Ativo  |
| POST   | `/patients/:id/labs/upload` | Ativo  |

## Listar labs (GET /v1/patients/:id/labs)

Parâmetros opcionais: `limit`, `offset`.

**Exemplo (curl):**
```bash
curl -i "https://api.sonnda.com.br/v1/patients/018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11/labs?limit=100&offset=0" \
  -H "Authorization: Bearer <id_token>"
```

## Listar labs completos (GET /v1/patients/:id/labs/full)

Inclui `test_results` com itens detalhados.

**Exemplo (curl):**
```bash
curl -i "https://api.sonnda.com.br/v1/patients/018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11/labs/full?limit=100&offset=0" \
  -H "Authorization: Bearer <id_token>"
```

## Upload de laudo (POST /v1/patients/:id/labs/upload)

Upload multipart com campo `file` (PDF/JPEG/PNG).

**Exemplo (curl):**
```bash
curl -i -X POST https://api.sonnda.com.br/v1/patients/018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11/labs/upload \
  -H "Authorization: Bearer <id_token>" \
  -F "file=@/caminho/para/laudo.pdf"
```

**Erros comuns:**
- `REQUIRED_FIELD_MISSING` (400) — arquivo ausente
- `UPLOAD_SIZE_EXCEEDED` (413) — arquivo muito grande
- `INVALID_FIELD_FORMAT` (400) — tipo não suportado
