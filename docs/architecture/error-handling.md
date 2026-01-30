<!-- docs/architecture/error-handling.md -->
# Error Handling (Go + Gin)

Este documento descreve a arquitetura de tratamento de erros da Sonnda API, inspirada em Hexagonal/Clean Architecture e aplicada de forma pragmática em Go.

**Referências**
- ADR: `docs/architecture/adr/ADR-002-error-handling-contrato.md`
- Catálogo de códigos: `internal/kernel/apperr/catalog.go`
- Política de log por erro: `internal/kernel/apperr/logging.go`
- Presenter HTTP (canonical): `internal/adapters/inbound/http/shared/httperr` (veja `APIErrorResponder`)
- Middleware de AccessLog: `internal/adapters/inbound/http/middleware/logging.go`
- Middleware de Recovery: `internal/adapters/inbound/http/middleware/recovery.go`

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

Implementado por `internal/adapters/inbound/http/shared/httperr.ToHTTP` + presenter `APIErrorResponder` no mesmo package.

---

## Fluxo por camada (visão geral)

```
HTTP Handler (Gin)
  ├─ valida/bind/parse (erros de fronteira)
  │    └─ cria *apperr.AppError { Code, Message, Cause }
  ├─ chama Service (internal/application/...)
  │    └─ Service converte erros esperados em *apperr.AppError
  └─ escreve resposta (internal/adapters/inbound/http/shared/httperr)
       ├─ ToHTTP(err) => (status, {code,message})
       └─ APIErrorResponder(c, err) | WebErrorResponder(c, err) => resposta + context keys + logging (5xx)
```

---

## Camadas e responsabilidades

### Domain (`internal/domain/...`)

- Retorna `error` (sentinelas/tipos) e valida invariantes.
- Pode fazer wrapping com `%w` para preservar a causa semântica sem “quebrar” `errors.Is`.
- Nunca importa HTTP e não conhece status codes.

### Application (`internal/application/...`)

Responsável por transformar erros “relevantes” em um erro de contrato: `*apperr.AppError`.

`*apperr.AppError` carrega:
- `Code` (`apperr.ErrorCode`) — contrato estável
- `Message` — seguro para o cliente
- `Cause` — erro interno (opcional), preservado via `Unwrap()` (suporta `errors.Is/As`)

Definição: `internal/kernel/apperr/error.go`

O catálogo de códigos fica em `internal/kernel/apperr/catalog.go`.

### HTTP adapter (presenter canonical)

O presenter canonical é o package `internal/adapters/inbound/http/shared/httperr` que exporta:
- `ToHTTP(err) (status int, body ErrorResponse)`
- `APIErrorResponder(c *gin.Context, err error)` — presenter para endpoints JSON (API)

Observação: implementação antiga em `internal/adapters/inbound/http/api/apierr` foi depreciada. Contribuições novas devem usar `httperr`.

---

## Mapeamento `code -> status`

O mapeamento é centralizado em `internal/adapters/inbound/http/shared/httperr/http_mapper.go` e segue as regras do catálogo `internal/kernel/apperr`.

---

## Padrão de uso nos handlers

Regras:
- Handler **não chama** `apperr.ToHTTP` diretamente.
- Handler deve chamar `internal/adapters/inbound/http/shared/httperr.APIErrorResponder(c, err)` para APIs JSON.
- Erros de fronteira (auth/bind/parse) são convertidos para `&apperr.AppError{Code, Message, Cause}` no handler e passados para o presenter.
- Erros do service/usecase **já devem** voltar como `*apperr.AppError` quando forem esperados.

Exemplo (padrão):

```go
import (
  httperr "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/httperr"
  "github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

if err := c.ShouldBindJSON(&req); err != nil {
  httperr.APIErrorResponder(c, &apperr.AppError{
    Code: apperr.VALIDATION_FAILED,
    Message: "payload inválido",
    Cause: err,
  })
  return
}

out, err := h.svc.Register(c.Request.Context(), input)
if err != nil {
  httperr.APIErrorResponder(c, err)
  return
}
```

---

## Checklist (novo code / novo erro)

1) Defina o `ErrorCode` em `internal/kernel/apperr/catalog.go`.
2) Garanta o status em `internal/adapters/inbound/http/shared/httperr/http_mapper.go`.
3) Defina o nível de log em `internal/kernel/apperr/logging.go` (se necessário).
4) No service, mapeie o erro do domínio para `*apperr.AppError` (ex.: `mapXDomainError`).
5) No handler, use `httperr.APIErrorResponder` ou `httperr.WebErrorResponder` (sem conversão manual para status/JSON).
6) Adicione/ajuste testes no service e (se fizer sentido) no presenter HTTP.
7) Atualize este documento e o ADR.
