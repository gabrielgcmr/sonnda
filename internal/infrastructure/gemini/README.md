<!-- internal/infrastructure/gemini/README.md -->
# Gemini - etapas 2.1 e 2.2

Cliente de transporte baseado em `google.golang.org/genai` v1.71.0.
Recebe texto e opcoes de geracao e devolve `genai.GenerateContentResponse`,
preservando candidatos, motivo de termino e consumo para o adaptador laboratorial.

## Configuracao

`config.Load()` disponibiliza `cfg.Gemini`. A chave pode ficar vazia enquanto
o runtime continua usando Document AI; `NewClient` exige uma chave explicita.
As configuracoes Google Cloud do storage nao substituem `GEMINI_API_KEY`.
O cliente seleciona explicitamente `BackendGeminiAPI`, sem alternar para Vertex AI.

| Variavel | Padrao | Finalidade |
| --- | --- | --- |
| `GEMINI_API_KEY` | Vazio | Credencial somente no backend |
| `GEMINI_MODEL` | `gemini-3.5-flash-lite` | Modelo candidato, configuravel |
| `GEMINI_TIMEOUT` | `30s` | Prazo por chamada, no formato de duracao Go |
| `GEMINI_MAX_INPUT_BYTES` | `131072` | Soma dos bytes UTF-8 de texto, instrucao e schema JSON |
| `GEMINI_MAX_OUTPUT_TOKENS` | `8192` | Limite de geracao enviado ao SDK |

Os limites sao defaults operacionais iniciais, nao limites oficiais do modelo.
Entrada acima do limite e rejeitada antes da chamada, sem truncamento.
O limite em bytes nao e uma contagem de tokens nem inclui o envelope HTTP do SDK.
Um deadline menor da requisicao e respeitado. Nao ha retry automatico nem fallback.

## Cliente

Criar `NewClient(ctx, cfg.Gemini)` uma vez e reutilizar o cliente.
O metodo `Generate` aceita `GenerateRequest` com `Text`, `SystemInstruction`
e `JSONSchema` opcional. Quando ha schema, configura `application/json` na resposta.
Nao ha ferramentas, busca externa nem upload de arquivos neste cliente.

## Extrator laboratorial

`NewLabReportTextExtractor(client)` implementa `labextraction.LabReportTextExtractor`.
Ele recebe `ExtractLabReportInput{Text: texto}`, chama o cliente Gemini com prompt
especifico para exames laboratoriais e valida a resposta contra o schema local de
`internal/domain/labextraction`.

O schema enviado ao provedor e derivado do schema local, removendo palavras-chave
de validacao que nao fazem parte do subconjunto suportado pelo SDK. O schema local
continua sendo a validacao final antes de converter a resposta para
`ExtractedLabReport`.

O adaptador trata:

- entrada vazia;
- erro do cliente;
- resposta sem candidato;
- resposta bloqueada;
- resposta truncada por limite de tokens;
- texto vazio;
- JSON invalido;
- JSON fora do schema;
- resposta valida sem itens, marcada como `needs_review`.

As rotas `/exames` e `/labs` continuam usando Document AI ate a etapa 4 da ADR-005.
Configurar a chave ou instanciar o cliente nao ativa o Gemini nessas rotas.
Nao registrar credenciais, texto clinico ou respostas completas nos logs comuns.

## Testes

`NewClientWithGenerator` recebe a interface `ContentGenerator`. Os testes injetam
uma implementacao simulada, sem credenciais reais nem acesso a rede.

```powershell
go test ./internal/config ./internal/infrastructure/gemini -v
```

Testes cobrem configuracao, limites, preservacao de texto/opcoes, timeout,
cancelamento, propagacao de erros, prompt, schema enviado ao provedor, validacao
local e interpretacao de respostas simuladas. A validacao com o modelo real esta
pendente da etapa 2.3; nenhum arquivo de `samples/` e enviado por esses testes.

## Avaliacao manual

A etapa 2.3 usa o comando local `cmd/lab-extraction-eval` para chamar o Gemini
real com arquivos `.txt`. Ele nao integra rotas nem banco.

```powershell
go run ./cmd/lab-extraction-eval -input samples/labs/glicose.txt
```

Com esperado:

```powershell
go run ./cmd/lab-extraction-eval -input samples/labs/glicose.txt -expected samples/labs/glicose.expected.json
```

Usar amostras sinteticas ou anonimizadas na avaliacao inicial.
