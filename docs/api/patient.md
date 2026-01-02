# Patient (create, update, soft delete)

Documentacao tecnica da implementacao de criacao, atualizacao e soft delete de pacientes.

## Visao geral (camadas)
- HTTP: `internal/http/api/handlers/patient_handler.go` (entrada API e parsing).
- Use cases: `internal/app/usecases/patientuc` (regras de negocio e orquestracao).
- Dominio: `internal/domain/entities/patient` (validacoes e normalizacao).
- Persistencia: `internal/infrastructure/persistence/supabase/patient_repo.go`.
- SQL: `internal/infrastructure/persistence/sqlc/sql/queries/patient_queries.sql`.

## Criar paciente
Arquivos-chave:
- Use case: `internal/app/usecases/patientuc/create_patient.go`.
- Handler: `internal/http/api/handlers/patient_handler.go` (CreatePatient).

Fluxo:
1) Handler valida autenticacao e faz bind do JSON.
2) Handler faz parsing/normalizacao de fronteira:
   - `birth_date` (layout `2006-01-02`).
   - `gender` e `race` via `shared.ParseGender/ParseRace`.
3) Use case monta entidade de dominio via `patient.NewPatient`.
   - Normaliza CPF, CNS, phone e avatar_url.
   - Valida campos obrigatorios (CPF, nome, data de nascimento).
4) Use case verifica duplicidade por CPF (`repo.FindByCPF`).
5) Persistencia grava com SQLC `CreatePatient`.
6) Retorna paciente criado.

Entradas relevantes:
- `cpf`, `full_name`, `birth_date`, `gender`, `race` (obrigatorios).
- `phone`, `avatar_url` (opcionais).

Erros principais:
- `ErrCPFAlreadyExists`, `ErrInvalidFullName`, `ErrInvalidBirthDate`.

## Atualizar paciente
Arquivos-chave:
- Use case: `internal/app/usecases/patientuc/update_patient.go`.
- Handler: `internal/http/api/handlers/patient_handler.go` (UpdateByID/UpdateByCPF).

Fluxo:
1) Handler autentica e faz bind do JSON em `PatientChanges`.
2) Use case busca paciente (por ID ou CPF).
3) Autorizacao (ReBAC): `authorization.PatientAuthorizer.Require` (membership usuario <-> paciente). Ver `docs/architecture/access-control.md`.
4) Entidade aplica mudancas via `ApplyUpdate`.
5) Persistencia grava com SQLC `UpdatePatient` (COALESCE para campos nulos).

Entradas permitidas:
- `full_name`, `phone`, `avatar_url`, `gender`, `race`, `cns`.

Erros principais:
- `ErrPatientNotFound`, `ErrAuthorizationForbidden`.

## Soft delete de paciente
Arquivos-chave:
- Use case: `internal/app/usecases/patientuc/delete_patient.go`.
- SQL: `SoftDeletePatient` em `internal/infrastructure/persistence/sqlc/sql/queries/patient_queries.sql`.

Fluxo:
1) Use case confirma existencia via `repo.FindByID`.
2) Persistencia executa soft delete (marca `deleted_at` e atualiza `updated_at`).

Comportamento esperado:
- Consultas de leitura (`GetPatientByID`, `GetPatientByCNS`, `ListPatients`, `SearchPatientsByName`)
  filtram `deleted_at IS NULL`.
- O soft delete preserva dados para auditoria e possivel restauracao.

Observacoes:
- Existe query `RestorePatient`, mas ainda nao esta exposta em use case/handler.
- Rotas de update/delete podem estar comentadas no router; caso habilite, verifique middleware de autenticacao/registro.
