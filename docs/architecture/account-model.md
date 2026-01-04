# Modelo de conta (AccountType + Professional.Kind)

Este documento descreve como o projeto modela "tipo de conta" e "tipo de profissional".

## Objetivo

Separar dois conceitos que antes ficavam misturados em um unico "role":

- **AccountType**: o tipo de conta no sistema (macro-permissoes; RBAC).
- **Professional.Kind**: o tipo de profissional (detalhe do profissional; usado em regras especificas).

Isso evita acoplar autorizacao a "ser paciente" e permite que um medico tambem seja paciente em outra relacao.

## Tipos

### AccountType (`internal/domain/model/user/account_type.go`)

- `professional`: conta de profissional de saude
- `basic_care`: conta de cuidado basico (ex.: caregiver)
- `admin`: reservado (fora do MVP; nao persistido no banco hoje)

Persistencia: `users.account_type` (PostgreSQL).

### Professional.Kind (`internal/domain/model/user/professional/kind.go`)

Catalogo de tipos de profissionais. Exemplos:

- `doctor`
- `nurse`
- `nursing_tech`
- `physiotherapist`
- `psychologist`
- `nutritionist`
- `pharmacist`
- `dentist`

Persistencia: `professionals.kind` (PostgreSQL).

## API (registro)

No `POST /api/v1/register`:

- sempre exige `account_type`
- se `account_type == professional`, exige `professional.kind` e os dados de registro (`registration_number`, `registration_issuer`, ...)

O handler faz parsing/validacao de fronteira e o service aplica regras e persiste.

Arquivos:

- Handler: `internal/http/api/handlers/user/user_handler.go`
- DTO inbound: `internal/app/ports/inbound/user/dto.go`
- Service: `internal/app/services/user/service_impl.go`

## Migracao

A mudanca de `users.role` para `users.account_type` e a adicao de `professionals.kind` foram registradas em:

`internal/infrastructure/persistence/sqlc/sql/migrations/0004_account_type_and_professional_kind.sql`

