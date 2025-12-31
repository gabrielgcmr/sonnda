# ADR-006 — Tratamento de erros baseado em Kind (apperr)

**Status:** Aceito  
**Data:** 2025-01  
**Contexto:** Sonnda API – tratamento de erros entre Domain/App e borda HTTP

---

## Contexto

Na Sonnda API, os módulos de domínio (`patient`, `user`, `professional`, `lab`, etc.) expõem erros próprios para sinalizar falhas de negócio, validação e autorização.

Inicialmente, a tradução desses erros para respostas HTTP era feita por meio de um `switch` centralizado na camada de handlers, comparando erros concretos de todos os domínios.

Esse modelo apresentou problemas claros:

- forte acoplamento entre a borda HTTP e todos os domínios
- crescimento contínuo de um `switch` global
- necessidade de alterar código HTTP sempre que um novo erro de domínio era criado
- dificuldade de reutilizar a mesma lógica fora do HTTP (ex.: jobs, CLI, gRPC)

Era necessário um mecanismo que separasse:
- **o significado do erro** (domínio/aplicação)
- **a forma como o erro é exposto externamente** (HTTP)

---

## Decisão

Foi adotado um modelo de erro tipado na camada de aplicação, denominado **`apperr`**, baseado em categorias semânticas (`Kind`).

### Estrutura principal

- Um erro da aplicação é representado por `apperr.Error`, que contém:
  - `Kind`: categoria semântica do erro
  - `Code`: identificador estável para clientes (ex.: `patient_not_found`)
  - `Message`: descrição humana (opcional)
  - `Err`: erro original encapsulado (opcional)

- Os domínios e services:
  - **não conhecem HTTP**
  - retornam erros com `Kind` apropriado
  - propagam erros sem traduzi-los

- A camada HTTP:
  - traduz **`apperr.Kind → decisão HTTP`**
  - centraliza essa lógica em um único ponto (`RespondAppError`)

---

## Kinds definidos

Os `Kind`s iniciais adotados foram:

- `not_found`
- `conflict`
- `invalid_input`
- `unauthorized`
- `forbidden`
- `external`
- `bad_gateway`
- `unavailable`
- `timeout`
- `service_closed`
- `internal`

Esses `Kind`s representam **significado semântico**, e não códigos HTTP diretamente.

---

## Implementação

### Criação do erro na aplicação/domínio

```go
var ErrPatientNotFound = apperr.New(
	apperr.KindNotFound,
	"patient_not_found",
	"patient not found",
)
