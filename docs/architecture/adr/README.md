
Este README referencia os ADRs quando necessário, mas **não substitui** esses registros.

---

## ADRs

- `docs/architecture/adr/ADR-005-repository-value-vs-service-pointer.md`
- `docs/architecture/adr/ADR-006-error-handling-contract.md`

---

## Padrões arquiteturais importantes

### Services retornam entidades completas

A camada de aplicação trabalha com **entidades completas do domínio** (`*entity`).

- Services não decidem o que será exposto publicamente
- Isso permite:
  - reutilização interna
  - auditoria
  - evolução do domínio sem quebrar contratos HTTP

A decisão sobre **quais dados retornam ao cliente** é responsabilidade da borda (HTTP).

---

### Handlers controlam exposição de dados

Handlers HTTP:
- definem explicitamente o payload de resposta
- podem retornar apenas `id + Location` em operações de criação
- mapeiam entidades para respostas públicas (DTOs ou payloads mínimos)

Esse padrão é fundamental para:
- segurança e LGPD
- controle de contratos públicos
- integração com builders de UI (ex.: Lovable)

---

### Repository retorna snapshots de persistência

Os repositórios seguem as seguintes regras:

- Finders unitários retornam `*Entity`
- Métodos de listagem retornam `[]Entity`
- Conversões para `[]*Entity` acontecem na camada de Service

Isso reforça a distinção entre:
- **snapshot de dados persistidos**
- **entidades vivas do domínio**

Além de evitar mutações acidentais e reduzir acoplamento.

---

### Separação clara entre leitura e mutação

- Operações de mutação (Create/Update/Delete) **não retornam objetos completos por padrão**
- Operações de leitura são explícitas (`GET /resource/:id`, listagens, etc.)

Esse padrão melhora:
- segurança
- previsibilidade da API
- ergonomia para clientes

---

## Quando criar um novo ADR

Um novo ADR deve ser criado quando:

- a decisão impactar múltiplos módulos
- a escolha não for óbvia para um novo contribuidor
- existirem alternativas plausíveis
- a decisão puder ser questionada no futuro

Exemplos comuns:
- ponteiros vs valores
- DTO vs entidade
- organização por feature vs camada
- contratos HTTP
- estratégias de autenticação/autorização

---

## Observação final

Este documento descreve **como a arquitetura está organizada**.

Os ADRs explicam **por que ela é assim**.

Sempre que uma decisão arquitetural relevante for tomada,  
o código e a documentação devem evoluir juntos.
