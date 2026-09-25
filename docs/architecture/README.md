<!-- docs/architecture/README.md -->
# Architecture

Descrição da arquitetura da Sonnda API, sua migração gradual por contexto, fluxos principais e decisões relevantes.

Este documento descreve **como a arquitetura está organizada**.  
As decisões não óbvias (o *porquê*) são registradas separadamente em ADRs.

---

## Visão geral

O backend está migrando de camadas globais para contextos em `internal/features`, mantendo a separação entre domínio, aplicação, HTTP e persistência. Os contextos ainda não migrados permanecem nas camadas existentes.

- **Domain (`internal/domain`)**  
  Modelos do domínio, regras de negócio e invariantes.  
  - Entities em `internal/domain/entity`; repositories em `internal/domain/repository`; storage abstractions em `internal/domain/storage`; contrato de extracao laboratorial em `internal/domain/labextraction`.

- **Application (`internal/application`)**  
  Orquestração e cross-cutting concerns.  
  - Use cases em `internal/application/usecase`; services em `internal/application/services` para os contextos ainda não migrados.
  - Bootstrapping (injeção de dependências) em `internal/application/bootstrap`.

- **API (`internal/api`)**  
  Implementações compartilhadas de adapters HTTP (inbound).  
  - Rotas em `internal/api/routes`; middlewares transversais em `internal/api/middleware`; presenter em `internal/api/presenter`.

- **Features (`internal/features`)**  
  Fluxos orientados a contexto de negócio.  
  - `auth` valida identidades externas e expõe `RequireBearer`.
  - `account` reúne serviços de perfil, onboarding, DTOs e mapeamento de erros.
  - `account/domain` contém `User`, `AccountType` e suas regras de validação, no pacote `accountdomain`.
  - `account/http` contém o handler de perfil e o middleware que resolve o usuário local e expõe `RequireRegisteredUser`.
  - `account/repository.go` define a interface de persistência; `account/postgres` implementa esse contrato usando o SQLC existente.
  - `patient/access` contém o checker, a listagem de pacientes acessíveis e os contratos de vínculo; seu handler HTTP atende `/v1/me/patients`.

- **Infrastructure (`internal/infrastructure`)**  
  Implementações concretas de persistência e integrações externas.  
  - **Persistence (`internal/infrastructure/persistence`)**: repositórios (sqlc/pgx), cache.
  - **Auth (`internal/infrastructure/auth`)**: Supabase auth provider.
  - **Document AI (`internal/infrastructure/documentai`)**: implementacao atual de extracao laboratorial via Google Cloud Document AI.

- **Kernel (`internal/kernel`)**  
  Preocupações transversais (cross-cutting concerns).  
  - Error contract (`internal/kernel/apperr`): `AppError` e catalog de códigos.
  - Observability (`internal/kernel/observability`): logging (slog) com escopo de requisição.
  - **Auth (`internal/infrastructure/auth`)**: integração com o provedor de autenticação.

- **Kernel (`internal/kernel`)**  
  Núcleo transversal do sistema.
  - **Error contract (`internal/kernel/apperr`)**: contrato centralizado de erros.
  - **Observability (`internal/kernel/observability`)**: logging baseado em slog, logger por request.

Essas camadas representam **limites conceituais**, não apenas organização de pastas.

## Migração de account — aplicação, persistência e entidades

```text
internal/features/account/
├── domain/
│   ├── user.go
│   ├── account_type.go
│   └── user_test.go
├── service.go
├── service_impl.go
├── dto.go
├── error_map.go
├── onboarding.go
├── onboarding_dto.go
├── repository.go
├── postgres/
│   ├── repository.go
│   └── repository_test.go
└── http/
    ├── handler.go
    ├── middleware.go
    └── middleware_test.go
```

O código antes distribuído entre `api/handlers/user.go`, `application/services/user`
e `application/usecase/registration` agora pertence a `account`. O pacote de
aplicação não depende de Gin; o transporte HTTP depende dos serviços de account e
dos helpers e presenter compartilhados.

`bootstrap/account.go` monta um único `AccountModule`, com handler e middleware
usando a mesma instância do repositório de usuários. `UserModule` foi incorporado
a esse módulo. As rotas continuam compostas em `internal/api/routes.go`.

A segunda etapa trouxe a interface `account.Repository` e o adaptador
`account/postgres.Repository` para a feature. O bootstrap fornece o pool ao
adaptador, que usa o SQLC já gerado, sem importar o pacote de repositórios legado.
Os erros de conflito e usuário ausente pertencem ao contrato de account. A falha
genérica de persistência pertence a `internal/domain/repository/errors.go`; o
pacote legado mantém uma referência ao mesmo erro para preservar a compatibilidade.
O mapeamento de erros da aplicação deixa de importar a implementação Postgres.

As entidades `User` e `AccountType`, seus parâmetros, erros de validação e testes
agora pertencem a `account/domain`, sem uma subpasta `entity`. Esse pacote mantém
as regras de domínio independentes de serviços de aplicação, HTTP, `apperr` e
infraestrutura. Os contextos de paciente importam `accountdomain` diretamente;
essa dependência entre contextos continua explícita, sem depender dos serviços
de account. A normalização de CPF continua usando o domínio compartilhado
`demographics`.

As interfaces e a listagem de acesso a pacientes pertencem a `patient/access`.
`account` não depende mais do repositório de acesso. A conexão compartilhada e o
código sqlc permanecem em infraestrutura. O onboarding não depende mais de um serviço ou perfil profissional
separado; registra os tipos de conta existentes pelo serviço de account. O
cadastro HTTP continua criando `basic_care`, conforme o contrato atual.
Não há migração de queries nesta etapa.

Não houve mudança de banco ou regras de concessão. O contrato da listagem deixa
de expor `relation_type`. Os testes
em `internal/api/account_routes_test.go` verificam os fluxos pelas rotas reais,
com serviços de account e repositórios em memória.
Os testes do adaptador verificam parâmetros, conversões e erros com uma
implementação em memória da interface de queries do SQLC, sem acessar banco real.

---

## Fluxo de request

1) **Middleware** autentica o usuário e adiciona informações ao contexto  
   (request_id, usuário autenticado, etc.).

2) **Handler HTTP**  
   - valida payload  
   - faz parsing de parâmetros  
   - monta o input do service  

3) **Service / Use case (camada App)**  
  - executa regras de negócio  
  - aplica políticas de acesso  
  - coordena chamadas a repositórios e serviços externos  

4) **Repository (Outbound)**  
  - executa queries via sqlc/pgx  
  - persiste ou consulta dados  

5) **Resposta HTTP**  
   - erros são normalizados para um contrato estável via `internal/kernel/apperr`
      - Veja `docs/architecture/error-handling.md`

---

## Persistência

- SQL definido em `internal/infrastructure/persistence/postgres/sqlc/sql`.
- `sqlc` gera código em `internal/infrastructure/persistence/postgres/sqlc/generated`.
- Repositórios de account ficam em `internal/features/account/postgres`; os demais continuam em `internal/infrastructure/persistence/postgres/repo`.
- Banco principal: PostgreSQL (Supabase).
- Soft delete usa `deleted_at`; consultas filtram `deleted_at IS NULL`.

---

## Observabilidade

- Logger baseado em `log/slog` (`internal/kernel/observability`).
- Variáveis:
  - `LOG_LEVEL`
  - `LOG_FORMAT`
- Um logger por request é injetado via middleware HTTP.

---

## Configuração

- Variáveis de ambiente definidas no ambiente (veja `.env.example` para referência).
- `APP_ENV` define o ambiente (`dev | prod`).
- Configurações carregadas na inicialização da aplicação (`internal/config`).

## OpenAPI

- `openapi.yaml` e `openapi/` são a fonte modular editável.
- `dist/openapi.yaml` é o bundle validado consumido por todos os geradores.
- `internal/generated/openapi` contém os tipos Go gerados.
- `internal/openapispec` contém os bytes do bundle servidos em `/openapi.yaml`.

---

## Bootstrap e rotas

- Bootstrap faz o wiring (repos, services e handlers) em `internal/application/bootstrap`.
- `AccountModule` reúne handler de perfil/onboarding e middleware de usuário registrado.
- As rotas HTTP vivem em `internal/api/routes.go` (API REST).  
- Níveis de acesso:
  - público
  - autenticado
  - registrado

---

## Decisões arquiteturais (ADR)

Algumas decisões importantes do projeto **não são óbvias apenas pela leitura do código**.  
Para preservar o contexto dessas escolhas ao longo do tempo, o Sonnda adota o uso de **Architecture Decision Records (ADR)**.

Os ADRs documentam:
- o contexto da decisão
- a decisão tomada
- alternativas consideradas
- consequências

Os ADRs vivem em:
`docs/architecture/adr/`.

---

## Controle de acesso aos pacientes

O pacote `internal/features/patient/access` centraliza a checagem de acesso por
vínculo. `RequireAccess` permite acesso ao dono do paciente ou
a um usuário com vínculo ativo; os demais recebem 403. Pacientes, exames e
laudos compartilham essa regra.

A mesma feature atende `GET /v1/me/patients`. A rota permanece estável, enquanto
o serviço, o handler e os DTOs deixam de pertencer a `account`.

As políticas por ação e profissão foram removidas, junto com a entidade, serviço
e repositório antigos de profissionais. `AccountType` permanece como dado da
conta e não concede acesso a pacientes. Tabelas, migrações e código SQLC gerado
foram preservados; sua limpeza é uma etapa separada.

Detalhes: `docs/architecture/access-control.md`.
