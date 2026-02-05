<!-- internal/api/openapi/README.md -->
# OpenAPI

Fonte de verdade do contrato HTTP da API.

- Spec: `internal/api/openapi/openapi.yaml` (embutido no binario e servido em `/openapi.yaml`).
- Codigo gerado: `internal/api/openapi/generated/oapi.gen.go`.
- Erros HTTP: RFC 9457 (Problem Details) via `application/problem+json`.

## Codegen

```bash
make oapi-codegen
```

## Validacao

```bash
make openapi-validate
```
