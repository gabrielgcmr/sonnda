# Controle de acesso (ReBAC)

Este projeto **nao** modela autorizacao como "RBAC + ABAC".

A direcao atual e **ReBAC** (Relationship-Based Access Control): o acesso e concedido a partir do **vinculo (relationship)** entre um app user e um recurso protegido (por enquanto, principalmente `patient`), e nao apenas por roles globais e nem por um grande conjunto de atributos da request.

## O que isso significa na pratica

- Um usuario nao "tem acesso a pacientes" globalmente.
- Um usuario tem acesso a **um paciente especifico** se existir um **vinculo ativo** entre eles.
- Esse vinculo pode carregar:
  - um **papel** no contexto daquele paciente (ex: caregiver, professional)
  - um conjunto de **permissoes** para acoes naquele paciente (ex: `patient:read`, `medical_record:labs:upload`)

Em codigo, esse vinculo e representado pela entidade `patientaccess.PatientAccess`.

## Modelo atual (MVP)

O modelo de autorizacao hoje gira em torno de:

- **Entidade de relacionamento:** `internal/domain/entities/patientaccess`
  - `PatientID`, `UserID` (quem acessa o que)
  - `Role` (papel no contexto do paciente)
  - `Permissions` (permissoes efetivas para este vinculo)
  - `Status/RevokedAt/ExpiresAt` existem no dominio (ainda evoluindo)

- **Authorizer (App):** `internal/app/services/authorization.PatientAuthorizer`
  - `Require(ctx, userID, patientID, perm)` aplica: "existe vinculo? esta ativo? inclui a permissao?"

- **Persistencia (hoje):** tabela `patient_access`
  - Schema: `internal/infrastructure/persistence/sqlc/sql/schema/patientaccess.sql`
  - Guarda `(patient_id, user_id, role, timestamps)`.
  - A persistencia de "lista de permissoes" e "status/expiracao" ainda esta sendo implementada ponta-a-ponta.

## Por que isso nao e RBAC + ABAC aqui

- **Nao e RBAC puro:** roles nao sao globais; `RoleProfessional` / `RoleCaregiver` em `patientaccess` sao roles *dentro do relacionamento*, nao "roles do sistema inteiro".
- **Nao e ABAC como base:** a decisao nao e feita principalmente por atributos como departamento, tenant, horario, device, etc. (podemos adicionar restricoes depois, mas a base continua relationship-first).

## Objetivos de design

- Proteger dados de paciente: negar por padrao e evitar vazar existencia de pacientes em acesso negado.
- Tornar compartilhamento explicito: acesso e criado/revogado gerenciando relacionamentos.
- Manter regras testaveis: a policy e aplicada via `PatientAuthorizer.Require` e a entidade `patientaccess`.

## Roadmap (ReBAC evoluindo)

ReBAC no projeto ainda esta em implementacao. Proximos passos comuns:

- modelar "invites" / "approvals" para criacao do vinculo
- persistir status/revogacao/expiracao no banco
- tratar "ownership" como relacionamento (ex: `patients.owner_user_id`) e definir permissoes implicitas
- suportar grafos mais ricos (time, delegacao, etc.) se necessario
