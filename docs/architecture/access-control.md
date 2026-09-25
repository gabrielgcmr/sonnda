<!-- docs/architecture/access-control.md -->
# Controle de acesso aos pacientes

O acesso atual verifica a quais pacientes uma conta está vinculada. Políticas
de autorização por ação, tipo de conta e profissão ficam para uma etapa futura.

A autenticação continua em `internal/features/auth`: valida a identidade externa.
O middleware de account resolve o cadastro local. O checker em
`internal/features/patient/access` recebe os identificadores da conta e do
paciente solicitado, sem depender de HTTP ou de um perfil profissional.

## Regra atual

`RequireAccess(ctx, accountID, patientID)` permite acesso quando a conta é
o dono (`OwnerUserID`) do paciente ou tem um vínculo ativo em `patient_access`.
Sem vínculo, a resposta é 403 (`ACCESS_DENIED`). Um paciente inexistente recebe
a mesma resposta, evitando revelar sua existência. Falhas de consulta não
concedem acesso e são retornadas pelo contrato central de erros.

A checagem é compartilhada pelo serviço de pacientes e pelos handlers de exames
e laudos, antes de ler, alterar ou processar dados. Ela é uma dependência
obrigatória desses fluxos. Os métodos de exclusão do serviço também exigem o
vínculo; não há mais bloqueio por ação. Nenhuma nova rota foi exposta.

A criação do paciente grava seu vínculo com o criador na mesma transação.
As listagens continuam usando a consulta de pacientes acessíveis ao usuário.
Ser médico ou ter `AccountType=professional` não concede acesso a outros pacientes.

## Modelos e persistência

- `internal/features/account/domain`: `User` e `AccountType`.
- `internal/features/patient/access/domain`: vínculo com o paciente e tipo de relacionamento.
- `internal/features/patient/access`: contratos de persistência e checker de acesso.
- `internal/features/patient/access/postgres`: adaptador PostgreSQL de acesso.

A entidade, o serviço e o repositório antigos de profissionais e as políticas
RBAC foram removidos. `AccountType` e o tipo de relacionamento permanecem como
dados existentes, sem políticas de permissão associadas. `relation_type` é
metadado do relacionamento e não concede ações. O contrato HTTP de
cadastro continua criando `basic_care`; ele não foi alterado nesta etapa.

Tabelas, migrações e código SQLC gerado de profissionais foram preservados.
A remoção desses artefatos de persistência deve ocorrer em uma etapa própria.

## Verificação

Os testes do checker de acesso cobrem dono, vínculo ativo, ausência de vínculo,
paciente inexistente, identidade ausente e falhas de consulta. Os testes dos
consumidores verificam que a negativa interrompe o fluxo antes de leituras,
alterações ou processamento de documentos.
