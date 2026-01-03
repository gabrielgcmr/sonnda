# Error Handling (Go + Gin)

Este documento descreve a arquitetura de tratamento de erros da Sonnda API, inspirada em Hexagonal/Clean Architecture e aplicada de forma pragmática em Go.

**Referências**
- ADR: `docs/architecture/adr/ADR-006-error-handling-contract.md`
- Catálogo de códigos: `internal/app/apperr/catalog.go`
- Política de log por erro: `internal/app/apperr/logging.go`
- Presenter HTTP: `internal/http/errors/error_presenter.go`
- Mapeamento HTTP: `internal/http/errors/http_mapper.go`
- Middleware de AccessLog: `internal/http/middleware/logging.go`
- Middleware de Recovery: `internal/http/middleware/recovery.go`

---

## Objetivos

- Definir **um contrato estável** de erro para clientes (frontend/mobile).
- Manter o **domínio agnóstico de HTTP** (sem status codes, sem Gin).
- Preservar causas técnicas com `%w` (observabilidade), mas **não vazar detalhes** na resposta HTTP.
- Evitar dependência circular entre pacotes.
- Evitar logs duplicados com baixo valor (1 log por request + 1 log detalhado só em 5xx).

---

## Contrato HTTP de erro

Formato padrão:

```json
{
  "error": { "code": "X", "message": "Y" }
}
```

- `code`: identificador estável (contrato com clientes).
- `message`: mensagem segura (não expõe detalhes internos).

Implementado por `internal/http/errors/ToHTTP` + `internal/http/errors/WriteError`.

---

## Fluxo por camada (visão geral)

```
HTTP Handler (Gin)
  ├─ valida/bind/parse (erros de fronteira)
  │    └─ cria *apperr.AppError { Code, Message, Cause }
  ├─ chama Service (internal/app/...)
  │    └─ Service converte erros esperados em *apperr.AppError
  └─ escreve resposta (internal/http/errors)
       ├─ ToHTTP(err) => (status, {code,message})
       └─ WriteError(c, err) => JSON + context keys + log (só 5xx)
```

---

## Camadas e responsabilidades

### Domain (`internal/domain/...`)

- Retorna `error` (sentinelas/tipos) e valida invariantes.
- Pode fazer wrapping com `%w` para preservar a causa semântica sem “quebrar” `errors.Is`, por exemplo:

```go
return "", fmt.Errorf("invalid race: %s: %w", input, shared.ErrInvalidRace)
```

- Nunca importa HTTP e não conhece status codes.

### Application (`internal/app/...`)

Responsável por transformar erros “relevantes” em um erro de contrato: `*apperr.AppError`.

`*apperr.AppError` carrega:
- `Code` (`apperr.ErrorCode`) — contrato estável
- `Message` — seguro para o cliente
- `Cause` — erro interno (opcional), preservado via `Unwrap()` (suporta `errors.Is/As`)

Definição: `internal/app/apperr/error.go`

O catálogo de códigos fica em `internal/app/apperr/catalog.go`:
- **AUTH**: `AUTH_REQUIRED`, `AUTH_TOKEN_INVALID`, `AUTH_TOKEN_EXPIRED`
- **AUTHZ**: `ACCESS_DENIED`, `ACTION_NOT_ALLOWED`
- **VALIDATION**: `VALIDATION_FAILED`, `REQUIRED_FIELD_MISSING`, `INVALID_FIELD_FORMAT`, `INVALID_ENUM_VALUE`, `INVALID_DATE`
- **NOT_FOUND**: `NOT_FOUND`
- **CONFLICT**: `RESOURCE_CONFLICT`, `RESOURCE_ALREADY_EXISTS`
- **DOMAIN**: `DOMAIN_RULE_VIOLATION` (422)
- **INFRA**: `INFRA_AUTHENTICATION_ERROR`, `INFRA_DATABASE_ERROR`, `INFRA_STORAGE_ERROR`, `INFRA_EXTERNAL_SERVICE_ERROR`, `INFRA_TIMEOUT`
- **RATE**: `RATE_LIMIT_EXCEEDED`, `UPLOAD_SIZE_EXCEEDED`
- **INTERNAL**: `INTERNAL_ERROR`

### HTTP adapter (`internal/http/...`)

Existe um pacote dedicado `internal/http/errors` (package name `errors`; normalmente importado como `httperrors` para evitar confusão com `errors` do stdlib).

Ele contém:

- `ToHTTP(err) (status int, body ErrorResponse)` em `internal/http/errors/http_response.go`
- `WriteError(c *gin.Context, err error)` em `internal/http/errors/error_presenter.go`

`WriteError`:
- chama `c.Error(err)` (registra no Gin)
- seta no `gin.Context`:
  - `error_code` (para AccessLog)
  - `http_status`
  - `error_log_level` (a partir de `apperr.LogLevelOf(err)`)
- escreve o JSON `{ "error": { "code", "message" } }`
- loga **detalhe do erro** (`slog.Any("err", err)`) **apenas em 5xx** (para reduzir ruído)

---

## Mapeamento `code -> status`

O mapeamento é centralizado em `internal/http/errors/http_mapper.go`:

| Categoria | Codes | Status |
|----------|-------|--------|
| AUTH | `AUTH_*` | 401 |
| AUTHZ | `ACCESS_DENIED`, `ACTION_NOT_ALLOWED` | 403 |
| VALIDATION | `VALIDATION_FAILED`, `REQUIRED_FIELD_MISSING`, `INVALID_FIELD_FORMAT`, `INVALID_ENUM_VALUE`, `INVALID_DATE` | 400 |
| NOT_FOUND | `NOT_FOUND` | 404 |
| CONFLICT | `RESOURCE_CONFLICT`, `RESOURCE_ALREADY_EXISTS` | 409 |
| DOMAIN | `DOMAIN_RULE_VIOLATION` | 422 |
| RATE | `RATE_LIMIT_EXCEEDED` / `UPLOAD_SIZE_EXCEEDED` | 429 / 413 |
| INFRA | `INFRA_*` | 5xx (500/502/504 conforme code) |
| INTERNAL | `INTERNAL_ERROR` | 500 |

---

## Padrão de uso nos handlers

Regras:
- Handler **não chama** `apperr.ToHTTP` diretamente.
- Handler chama `httperrors.WriteError(c, err)` para qualquer erro a ser respondido.
- Erros de fronteira (auth/bind/parse) são convertidos para `&apperr.AppError{Code, Message, Cause}` no handler e passados para `WriteError`.
- Erros do service/usecase **já devem** voltar como `*apperr.AppError` quando forem esperados.

Exemplo (padrão):

```go
import (
  httperrors "sonnda-api/internal/http/errors"
  "sonnda-api/internal/app/apperr"
)

if err := c.ShouldBindJSON(&req); err != nil {
  httperrors.WriteError(c, &apperr.AppError{
    Code: apperr.VALIDATION_FAILED,
    Message: "payload inválido",
    Cause: err,
  })
  return
}

out, err := h.svc.Register(c.Request.Context(), input)
if err != nil {
  httperrors.WriteError(c, err)
  return
}
```

---

## Como criar `AppError` no service

Padrão recomendado: mapear erros do domínio/infra para `*apperr.AppError` em funções auxiliares por service/feature (mantém o `service_impl.go` limpo).

Exemplos reais:
- `internal/app/services/user/error.go` (`mapUserDomainError`, `mapInfraError`)
- `internal/app/services/patient/error.go` (`mapPatientDomainError`, `mapInfraError`)

---

## Logging policy (AccessLog vs Writer)

### AccessLog (`internal/http/middleware/logging.go`)

Sempre gera 1 log por request com:
- `request_id`, `status`, `latency_ms`, `method`, `path`, `route`, `client_ip`, `response_bytes`, `user_agent`
- inclui `error_code` se existir
- respeita `error_log_level` (se existir) para decidir nível em `request_invalid`

### Writer (`internal/http/errors/error_presenter.go`)

- Para **4xx**: não loga detalhe (evita ruído); AccessLog já registra o request.
- Para **5xx**: faz 1 log detalhado (`handler_error`) com `err` e contexto.

### Recovery middleware (`internal/http/middleware/recovery.go`)

Captura `panic`, loga stacktrace e responde `500` com payload genérico (`internal_error`), usando o logger contextual quando disponível.

---

## Anti‑padrões

- Usar `err.Error()` como contrato público (quebra clientes e vaza detalhes).
- Handlers mapeando dezenas de erros do domínio diretamente (espalha regra e gera inconsistência).
- Logar o mesmo erro 2–3 vezes por request (sem agregar valor).
- Domínio importando Gin/HTTP/status codes.

---

## Checklist (novo code / novo erro)

1) Defina o `ErrorCode` em `internal/app/apperr/catalog.go`.
2) Garanta o status em `internal/http/errors/http_mapper.go`.
3) Defina o nível de log em `internal/app/apperr/logging.go` (se necessário).
4) No service, mapeie o erro do domínio para `*apperr.AppError` (ex.: `mapXDomainError`).
5) No handler, use `httperrors.WriteError(c, err)` (sem conversão manual para status/JSON).
6) Adicione/ajuste testes no service e (se fizer sentido) no presenter HTTP.
7) Atualize este documento e o ADR.

