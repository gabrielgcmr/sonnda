<!-- internal/infrastructure/gemini/README.md -->
# Cliente Gemini - etapa 2.1

Cliente de transporte baseado em `google.golang.org/genai` v1.71.0.
Recebe texto e opcoes de geracao e devolve `genai.GenerateContentResponse`,
preservando candidatos, motivo de termino e consumo para a etapa seguinte.

## Configuracao

`config.Load()` disponibiliza `cfg.Gemini`. A chave pode ficar vazia enquanto
o runtime continua usando Document AI; `NewClient` exige uma chave explicita.
As configuracoes Google Cloud do storage nao substituem `GEMINI_API_KEY`.
O cliente seleciona explicitamente `BackendGeminiAPI`, sem alternar para Vertex AI.

| Variavel | Padrao | Finalidade |
| --- | --- | --- |
| `GEMINI_API_KEY` | Vazio | Credencial somente no backend |
| `GEMINI_MODEL` | `gemini-2.5-flash-lite` | Modelo candidato, configuravel |
| `GEMINI_TIMEOUT` | `30s` | Prazo por chamada, no formato de duracao Go |
| `GEMINI_MAX_INPUT_BYTES` | `131072` | Soma dos bytes UTF-8 de texto, instrucao e schema JSON |
| `GEMINI_MAX_OUTPUT_TOKENS` | `8192` | Limite de geracao enviado ao SDK |

Os limites sao defaults operacionais iniciais, nao limites oficiais do modelo.
Entrada acima do limite e rejeitada antes da chamada, sem truncamento.
O limite em bytes nao e uma contagem de tokens nem inclui o envelope HTTP do SDK.
Um deadline menor da requisicao e respeitado. Nao ha retry automatico nem fallback.

## Uso na proxima etapa

Criar `NewClient(ctx, cfg.Gemini)` uma vez e reutilizar o cliente.
O metodo `Generate` aceita `GenerateRequest` com `Text`, `SystemInstruction`
e `JSONSchema` opcional. Quando ha schema, configura `application/json` na resposta.
Nao ha ferramentas, busca externa nem upload de arquivos neste cliente.

O schema e encaminhado ao SDK; compatibilidade com o fornecedor e validacao local
completa serao implementadas na etapa 2.2. O cliente nao converte a resposta em
`ExtractedLabReport` nem interpreta bloqueios, truncamento ou ausencia de resultados.
Portanto, ainda nao implementa `LabReportTextExtractor`: essa responsabilidade sera
do adaptador laboratorial, sem criar um metodo temporario que apenas devolve erro.

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
cancelamento e propagacao de erros. A validacao com o modelo real esta pendente
das etapas 2.2 e 2.3; nenhum arquivo de `samples/` e enviado por esses testes.
