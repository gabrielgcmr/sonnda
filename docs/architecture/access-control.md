# Controle de acesso (RBAC + ReBAC)

Este projeto nao modela autorizacao como "RBAC" puro.

O modelo atual combina:

- **RBAC (acoes)** para limitar *o que* um tipo de conta pode fazer (ex.: "basic_care pode ler paciente?", "professional pode fazer upload de labs?").
- **ReBAC (relacionamento)** para limitar *em qual paciente* o usuario pode operar (ex.: "tem vinculo com este paciente? e dono?").

O objetivo e manter:
- contrato simples e estavel para o cliente
- dominio agnostico de HTTP
- politica de acesso testavel e centralizada na camada App (authorizer, a ser criado)

## O que isso significa na pratica

- Um usuario nao "tem acesso a pacientes" globalmente.
- Um usuario tem acesso a **um paciente especifico** se:
  - ele e o **owner** (`patients.owner_user_id`), ou
  - existir um **vinculo** em `patient_access (patient_id, user_id, role)`.
- "Ser paciente" nao e uma role global: `patient` e um recurso/registro, e pode ou nao ter uma conta vinculada.

Em codigo, o vinculo e representado no dominio por `internal/domain/model/patient/patientaccess.PatientAccess`.

## Modelo atual (MVP)

O modelo de autorizacao hoje gira em torno de:

- **Tipos de conta (RBAC):** `internal/domain/model/user.AccountType`
  - `professional`: conta de profissional de saude
  - `basic_care`: conta de cuidado basico (ex.: caregiver)
  - `admin`: reservado (fora do MVP; nao persistido no banco hoje)

- **Tipo de profissional (apenas se AccountType==professional):** `internal/domain/model/user/professional.Kind`
  - exemplos: `doctor`, `nurse`, ...
  - persistido em `professionals.kind`

- **RBAC (acoes):** `internal/domain/model/rbac`
  - `rbac.Action` e o catalogo de acoes (`patient:read`, `labs:upload`, etc.)
  - `rbac.Subject` representa "quem e o ator" (AccountType + optional ProfessionalKind)
  - `rbac.RbacPolicy.CanPerform(subject, action)` decide se a acao e permitida *em abstrato*

- **Entidade de relacionamento (ReBAC):** `internal/domain/model/patient/patientaccess`
  - `PatientID`, `UserID` (quem acessa o que)
  - `RelationType` (papel no contexto do paciente: caregiver, professional, self, etc.)
  - `RevokedAt` existe no dominio (no banco, revoke hoje e feito por delete)

- **Persistencia (hoje):**
  - `users.account_type` (substitui `users.role`)
  - `professionals.kind`
  - `patient_access.role` (papel do vinculo no contexto do paciente)
  - schema principal em `internal/infrastructure/persistence/sqlc/sql/schema/users.sql` e `internal/infrastructure/persistence/sqlc/sql/schema/patientaccess.sql`
  - migracao: `internal/infrastructure/persistence/sqlc/sql/migrations/0004_account_type_and_professional_kind.sql`

## Por que isso nao e RBAC + ABAC aqui

- **Nao e RBAC puro:** roles nao sao globais; `RoleProfessional` / `RoleCaregiver` em `patientaccess` sao roles *dentro do relacionamento*, nao "roles do sistema inteiro".
- **Nao e ABAC como base:** a decisao nao e feita principalmente por atributos como departamento, tenant, horario, device, etc. (podemos adicionar restricoes depois, mas a base continua relationship-first).

O RBAC aqui existe para limitar **acoes** por tipo de conta/profissional, mas nunca substitui o cheque de relacionamento com o paciente.

## Objetivos de design

- Proteger dados de paciente: negar por padrao e evitar vazar existencia de pacientes em acesso negado.
- Tornar compartilhamento explicito: acesso e criado/revogado gerenciando relacionamentos.
- Manter regras testaveis: RBAC no dominio (`rbac.RbacPolicy`) e ReBAC na camada App (authorizer a ser criado) usando repositorios.

## Roadmap (ReBAC evoluindo)

ReBAC no projeto ainda esta em implementacao. Proximos passos comuns:

- modelar "invites" / "approvals" para criacao do vinculo
- persistir status/revogacao/expiracao no banco
- tratar "ownership" como relacionamento (ex: `patients.owner_user_id`) e definir permissoes implicitas
- suportar grafos mais ricos (time, delegacao, etc.) se necessario
