# Patient (create, get, update, soft delete)

Documentacao tecnica da implementacao do modulo de pacientes.

## Visao geral (camadas)

- HTTP: `internal/http/api/handlers/patient/patient_handler.go` (entrada API, bind/parse e resposta).
- Application (service): `internal/app/services/patient` (orquestracao e mapeamento de erros para `*apperr.AppError`).
- Dominio: `internal/domain/model/patient` (entidade, invariantes e validacoes).
- Persistencia: `internal/infrastructure/persistence/repository/patient/patient_repo.go`.
- SQL (sqlc): `internal/infrastructure/persistence/sqlc/sql/queries/patient_queries.sql`.

## Status atual

- Criacao e busca por ID estao implementadas ponta-a-ponta.
- List/update/soft delete existem no service/handler, mas podem depender de implementacao completa no repository e de rotas habilitadas no router.

## Criar paciente (POST /api/v1/patients)

Arquivos-chave:
- Handler: `internal/http/api/handlers/patient/patient_handler.go` (Create)
- Service: `internal/app/services/patient/service_impl.go` (Create)
- Dominio: `internal/domain/model/patient/patient.go` (NewPatient)
- Repository: `internal/infrastructure/persistence/repository/patient/patient_repo.go` (Create, FindByCPF)

Fluxo (alto nivel):
1) Handler valida autenticacao/registro e faz bind do JSON.
2) Handler faz parsing/normalizacao de fronteira:
   - `birth_date` (layout `2006-01-02`) via `common.ParseBirthDate`.
   - `gender` e `race` via `ParseGender/ParseRace` do handler.
3) Service monta entidade via `patient.NewPatient` (normaliza/valida invariantes).
4) Service checa duplicidade por CPF (`repo.FindByCPF`).
5) Repository persiste via sqlc (`CreatePatient`).

Erros principais (contrato HTTP):
- `VALIDATION_FAILED` (400): payload invalido / dados invalidos
- `RESOURCE_ALREADY_EXISTS` (409): CPF ja cadastrado
- `INFRA_DATABASE_ERROR` (5xx): falha tecnica (banco)

## Buscar paciente por ID (GET /api/v1/patients/:id)

Arquivos-chave:
- Handler: `internal/http/api/handlers/patient/patient_handler.go` (GetByID)
- Service: `internal/app/services/patient/service_impl.go` (GetByID)
- Repository: `internal/infrastructure/persistence/repository/patient/patient_repo.go` (FindByID)

Fluxo (alto nivel):
1) Handler valida autenticacao/registro.
2) Handler valida `:id` e faz `uuid.Parse`.
3) Service aplica policy de acesso e busca no repo (`FindByID`).
4) Se nao encontrar, retorna `NOT_FOUND` (404).

## Listar pacientes (GET /api/v1/patients)

- Handler: `internal/http/api/handlers/patient/patient_handler.go` (List)
- Service: `internal/app/services/patient/service_impl.go` (List)

Notas:
- Atualmente o handler usa `limit=100` e `offset=0` (sem paginacao via querystring).

## Atualizar paciente (PUT /api/v1/patients/:id)

- Handler existe em `internal/http/api/handlers/patient/patient_handler.go` (UpdateByID).
- Service existe em `internal/app/services/patient/service_impl.go` (UpdateByID).
- A rota pode estar comentada em `internal/http/api/router.go`.

## Soft delete de paciente

- Service existe em `internal/app/services/patient/service_impl.go` (SoftDeleteByID).
- A rota pode estar comentada em `internal/http/api/router.go`.
