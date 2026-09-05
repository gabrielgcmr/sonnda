<!-- internal/domain/labextraction/README.md -->
# Contrato de extracao laboratorial por texto

Etapa 1 da ADR-005. Este pacote define a entrada textual, a saida existente em Go
e o JSON Schema versionado. Ainda nao chama Gemini nem muda o upload atual.

## Entrada e transicao

`ExtractLabReportInput{Text: texto}` aceita texto nao vazio e preserva a entrada.
`LabReportTextExtractor` e a interface alvo para a etapa 2.
Paciente, usuario, URI, IDs e persistencia continuam sob responsabilidade do caso de uso.

`LabReportExtractor` ainda recebe URI e MIME type para manter Document AI funcionando.
Essa compatibilidade temporaria deve ser removida na etapa 4, quando as duas rotas
passarem a fornecer texto. O novo contrato nao e fallback do antigo.

## Correspondencia com o banco

| JSON | Destino existente |
| --- | --- |
| `patient_name`, `patient_dob`, `lab_name`, `lab_phone` | Campos correspondentes em `lab_reports` |
| `insurance_provider`, `requesting_doctor`, `technical_manager`, `report_date` | Campos correspondentes em `lab_reports` |
| `tests[]` | Registros em `lab_results` |
| `test_name`, `material`, `method`, `collected_at`, `release_at` | Campos correspondentes em `lab_results` |
| `tests[].items[]` | Registros em `lab_result_items` |
| `parameter_name`, `result_value`, `result_unit`, `reference_text` | Campos correspondentes em `lab_result_items` |

O schema acompanha os campos de extracao usados pelo mapper atual. Nao e um dump
das tabelas: IDs, chaves estrangeiras, fingerprint, timestamps de gravacao, texto
bruto e metadados tecnicos nao sao gerados pelo modelo.
`RawText` e os metadados continuam existindo no DTO para preenchimento pelo backend,
mas sao excluidos do JSON de resposta da LLM.

## Regras do JSON

- Arquivo: `lab_report.schema.json`, acessivel em Go por `LabReportResponseSchema()`.
- Versao: `LabReportSchemaVersion`. Mudancas incompativeis exigem revisar a versao.
- As chaves declaradas sao obrigatorias; campos opcionais clinicamente usam `null`.
- `result_value` e texto, como no banco: `14,2`, `245.000`, `< 5`, `Negativo`.
- Nao converter unidades, preencher valores ausentes ou criar nomes canonicos.
- Referencias permanecem compativeis, mas podem ser `null` na primeira versao.
- Datas seguem as representacoes descritas no schema e aceitas pelo mapper atual.
  A validacao de calendario e plausibilidade nao faz parte destes testes de formato.
- Percentual e absoluto sao itens distintos, preservando nome, valor e unidade.
- `tests: []` representa ausencia de resultados. Um pedido ou atestado sem resultados
  nao deve gerar itens somente por mencionar analitos, datas ou duracoes.
- Um item com valor ilegivel pode usar `result_value: null`; o adaptador futuro
  devera sinalizar essa incompletude. Schema valido nao significa sucesso clinico.

As descricoes do schema orientam a extracao, mas nao sao validacoes semanticas.
`HasStructuredResults()` atualmente verifica nomes de exame e analito; nao garante
valor preenchido. O adaptador da etapa 2 devera validar itens utilizaveis antes de
declarar sucesso. Este JSON Schema nao valida datas reais nem fidelidade ao documento.

## Exemplos e samples

`testdata/` contem exemplos sinteticos versionados. Cada caso tem:

- `<nome>.input.txt`: texto de entrada;
- `<nome>.expected.json`: saida esperada, revisada manualmente;
- entrada em `cases.json`: nome e indicacao de presenca de resultados estruturados.

Os testes validam o schema e a compatibilidade do JSON com os DTOs. Nao convertem
o texto em JSON e nao comprovam que um modelo reconhece pedidos ou atestados.
Na etapa 2, estes pares servirao para comparar a resposta real do extrator.

A pasta `samples/` na raiz da API contem os arquivos locais para avaliacao manual
e esta ignorada pelo Git. Ela nao e lida nem enviada a servicos por estes testes.
Para transformar um caso em teste versionado, criar um exemplo sintetico ou
anonimizado em `testdata/`, revisar a saida esperada e registrar em `cases.json`.
Nao basta copiar o JSON produzido pela propria LLM como resposta correta.

```powershell
go test ./internal/domain/labextraction -v
```

O validador JSON Schema usado pelos testes e `github.com/santhosh-tekuri/jsonschema/v6`.
A validacao do schema funciona localmente, sem API key, OCR ou banco de dados.
