<!-- docs/architecture/adr/ADR-005-extracao-semantica-exames-laboratoriais.md -->
# ADR-005 - Extracao semantica para exames laboratoriais

**Status:** Aceito  
**Data:** 2026-09-05  
**Contexto:** Sonnda API - extracao estruturada de exames laboratoriais

---

## Contexto

A aplicacao precisa receber exames laboratoriais e disponibilizar os dados mais importantes para o paciente: nome do exame, analitos, valores, unidades, datas e, quando possivel, referencias.

A ADR-004 propunha parser deterministico para exames laboratoriais comuns. A implementacao inicial mostrou que mesmo casos aparentemente simples, como hemograma e glicose/glicemia, exigem muitas regras especificas, aliases, tratamento de layout e excecoes de OCR.

Essa complexidade tende a crescer muito com a variedade real de exames laboratoriais:

- lipidograma;
- hormonios;
- vitaminas;
- enzimas hepaticas;
- urina/EAS;
- culturas;
- sorologias;
- exames com resultados qualitativos;
- diferentes laboratorios e layouts.

Manter um parser deterministico por exame como estrategia principal criaria alto custo de manutencao e risco de codigo morto ou incompleto.

---

## Decisao

Usar extracao semantica estruturada como estrategia principal para exames laboratoriais.

O pipeline alvo passa a ser:

```text
arquivo
  -> upload seguro
  -> extracao de texto/OCR quando aplicavel
  -> extrator semantico com schema estruturado
  -> normalizacao deterministica leve
  -> validacao conservadora
  -> persistencia dos resultados laboratoriais
  -> exibicao estruturada no cliente
```

O extrator semantico pode ser Document AI, LLM com schema ou outro servico equivalente. A decisao arquitetural nao fica presa a um fornecedor especifico; o contrato interno deve ser um schema de saida estavel.

Parsers deterministos nao serao usados como estrategia principal. Eles podem existir apenas como utilitarios pequenos e comprovadamente vantajosos, por exemplo:

- normalizar numero brasileiro;
- validar unidade;
- validar plausibilidade tecnica ampla;
- limpar texto bruto;
- checar consistencia de campos ja extraidos.

---

## Regras de desenho

### Schema estruturado como contrato interno

A saida esperada da extracao laboratorial deve seguir um contrato do tipo:

```go
type ExtractedLabReport struct {
    PatientName      *string
    LabName          *string
    ReportDate       *string
    RawText          *string
    Tests            []ExtractedLabTest
}

type ExtractedLabTest struct {
    TestName     string
    Material     *string
    Method       *string
    CollectedAt  *string
    ReleaseAt    *string
    Items        []ExtractedLabItem
}

type ExtractedLabItem struct {
    ParameterName string
    ResultValue   *string
    ResultUnit    *string
    ReferenceText *string
}
```

O schema pode evoluir, mas deve continuar priorizando dados estruturados em vez de texto corrido.

### Texto bruto continua sendo preservado

Mesmo usando extracao semantica, o texto bruto deve ser mantido quando disponivel.

Ele serve para:

- auditoria;
- debug de extracao;
- reprocessamento futuro;
- fallback de leitura humana.

### Validacao deterministica apos extracao

Regras deterministicas continuam uteis depois da extracao, mas como camada de controle, nao como parser principal.

Exemplos:

- converter `15,1` para `15.1` quando necessario;
- marcar resultado suspeito quando um valor estiver fora de faixa tecnica ampla;
- detectar item sem unidade quando unidade for esperada;
- preservar item como pendente de revisao quando a confianca for baixa.

### Falha parcial e aceitavel

Falhas em alguns itens nao devem invalidar o documento inteiro.

O sistema deve aceitar estados como:

- extraido com sucesso;
- extraido parcialmente;
- precisa revisao;
- falhou.

### Evitar codigo morto

Implementacoes experimentais que nao forem integradas ao fluxo principal devem ser removidas ou movidas explicitamente para uma area experimental documentada.

Para esta decisao, o pacote `internal/domain/labparser` deve ser removido do runtime principal.

---

## Alternativas consideradas

### Parser deterministico por exame

Substituido.

Era atraente por custo baixo e auditabilidade, mas mostrou alto custo de manutencao e baixa escalabilidade diante da variedade real de exames.

### Salvar apenas texto corrido

Rejeitado para laboratorio.

Texto corrido pode ser suficiente para alguns laudos de imagem, mas exames laboratoriais precisam de valores estruturados para comparacao e acompanhamento.

### Usar extracao semantica sem validacao

Rejeitado.

Mesmo com extrator semantico, a aplicacao deve preservar texto bruto e aplicar validacoes conservadoras para reduzir erro silencioso.

---

## Consequencias

### Positivas

- Escala melhor para muitos tipos de exames.
- Reduz manutencao manual de regex e aliases por exame.
- Mantem a experiencia principal: valores laboratoriais estruturados.
- Permite trocar fornecedor de extracao mantendo contrato interno.
- Mantem validacao deterministica onde ela agrega valor.

### Negativas / trade-offs

- Maior custo por processamento em comparacao com parser local.
- Dependencia de servico externo ou modelo semantico.
- Exige bons testes de contrato e observabilidade para avaliar qualidade.
- Pode precisar de fila/reprocessamento quando o servico externo falhar.

---

## Plano de evolucao

1. Remover o pacote experimental `internal/domain/labparser`.
2. Consolidar o contrato interno de extracao laboratorial.
3. Garantir que `POST /v1/patients/:id/exames` usa o mesmo caminho estruturado de labs para documentos laboratoriais.
4. Persistir vinculo entre `exam_documents`, `exam_document_texts` e `lab_reports`.
5. Exibir no Flutter os itens laboratoriais estruturados quando existirem.
6. Adicionar validacoes deterministicas leves apos a extracao.
7. Adicionar observabilidade para taxa de sucesso, falha parcial e documentos que precisam revisao.
