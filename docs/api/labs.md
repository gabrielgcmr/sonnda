<!-- docs/api/labs.md -->
# Labs

Endpoints de laudos laboratoriais.

## Base URL

`https://api.sonnda.com.br/v1`

## Autenticação

Todas as rotas exigem `Authorization: Bearer <id_token>`.

## Contrato oficial

O contrato completo de endpoints, schemas e erros fica no OpenAPI: `/openapi.yaml`.

## Listar labs (GET /v1/patients/:id/labs)

Parâmetros opcionais: `limit`, `offset`, `expand`, `include`.

- `expand=full` retorna a representação completa.
- `include=results` é equivalente a `expand=full`.

**Exemplo (curl):**
```bash
curl -i "https://api.sonnda.com.br/v1/patients/018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11/labs?limit=100&offset=0" \
  -H "Authorization: Bearer <id_token>"
```

## Upload de laudo (POST /v1/patients/:id/labs)

Upload multipart com campo `file` (PDF/JPEG/PNG).

**Exemplo (curl):**
```bash
curl -i -X POST https://api.sonnda.com.br/v1/patients/018f3a2a-4c1a-7c5a-9d9e-2b7d8d9c3f11/labs \
  -H "Authorization: Bearer <id_token>" \
  -F "file=@/caminho/para/laudo.pdf"
```

**Dicas:**
- `expand=full` e `include=results` retornam a representação completa.
