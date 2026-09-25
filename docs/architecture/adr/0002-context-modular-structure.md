<!-- docs/architecture/adr/0002-context-modular-structure.md -->
# ADR-0002: Organização modular por contexto

## Status
Accepted

## Contexto
A estrutura atual em camadas funciona, mas alguns componentes de API e bootstrap cresceram em arquivos globais por tipo técnico, dificultando evolução do domínio clínico.

## Decisão
Adotar organização por contexto de negócio, preservando as camadas:
- `account`
- `patient`
- `medicalrecord`

`medicalrecord` é tratado como agregado raiz para capacidades clínicas como exames, problemas e medicações.

## Regras
- Contratos e wiring devem nascer no contexto correspondente.
- `kernel` permanece exclusivo para cross-cutting (erro, observabilidade, auth comum).
- Novas funcionalidades clínicas devem ser adicionadas sob `medicalrecord/*`.

## Consequências
- Melhora a navegabilidade e reduz acoplamento acidental.
- Permite migração incremental sem quebra imediata de comportamento.
- Requer atualização gradual de imports e módulos existentes para a estrutura contextual.
