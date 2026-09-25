<!-- docs/architecture/access-control.md -->
# Controle de acesso aos pacientes

A autorização atual verifica em quais pacientes o usuário pode operar. Políticas
por ação, tipo de conta e profissão ficam para uma etapa futura.

A autenticação continua em `internal/features/auth`: valida a identidade externa.
O middleware de account resolve o cadastro local. O pacote
`internal/application/services/authorization` recebe esse usuário e o paciente
solicitado, sem depender de HTTP ou de um perfil profissional.

## Regra atual

`RequirePatientAccess(ctx, actor, patientID)` permite acesso quando o usuário é
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
- `internal/domain/entity/patientaccess`: vínculo com o paciente.
- `internal/domain/repository`: interfaces de pacientes e vínculos.
- `internal/application/services/authorization`: validação compartilhada do acesso.

A entidade, o serviço e o repositório antigos de profissionais e as políticas
RBAC foram removidos. `AccountType` e o tipo de relacionamento permanecem como
dados existentes, sem políticas de permissão associadas. O contrato HTTP de
cadastro continua criando `basic_care`; ele não foi alterado nesta etapa.

Tabelas, migrações e código SQLC gerado de profissionais foram preservados.
A remoção desses artefatos de persistência deve ocorrer em uma etapa própria.

## Verificação

Os testes do autorizador cobrem dono, vínculo ativo, ausência de vínculo,
paciente inexistente, identidade ausente e falhas de consulta. Os testes dos
consumidores verificam que a negativa interrompe o fluxo antes de leituras,
alterações ou processamento de documentos.
