<!-- internal/infrastructure/gemini/README.md -->
# Gemini - Extrator Laboratorial

Cliente de transporte baseado em `google.golang.org/genai`.
Recebe texto e opções de geração e devolve `genai.GenerateContentResponse`,
preservando candidatos, motivo de término e consumo para o adaptador laboratorial.

## Configuração

`config.Load()` disponibiliza `cfg.Gemini`.
`NewClient` exige chave explícita (`GEMINI_API_KEY`).
O cliente seleciona explicitamente `BackendGeminiAPI`, sem alternar para Vertex AI.

| Variável | Padrão | Finalidade |
| --- | --- | --- |
| `GEMINI_API_KEY` | Vazio | Credencial somente no backend |
| `GEMINI_MODEL` | `gemini-2.5-flash` | Modelo configurável |
| `GEMINI_TIMEOUT` | `30s` | Prazo por chamada, no formato de duração Go |
| `GEMINI_MAX_INPUT_BYTES` | `131072` | Limite de bytes UTF-8 de entrada |
| `GEMINI_MAX_OUTPUT_TOKENS` | `8192` | Limite de geração enviado ao SDK |

Um deadline menor da requisição é respeitado. Não há retry automático nem fallback.

## Cliente

Criar `NewClient(ctx, cfg.Gemini)` uma vez e reutilizar o cliente.
O método `Generate` aceita `GenerateRequest` com `Text`, `SystemInstruction`
e `JSONSchema` opcional. Quando há schema, configura `application/json` na resposta.

## Extrator laboratorial

`NewLabReportTextExtractor(client)` implementa `labextraction.LabReportTextExtractor`.
Ele recebe `ExtractLabReportInput{Text: texto}`, chama o cliente Gemini com prompt
específico para exames laboratoriais e valida a resposta contra o schema local de
`internal/features/documentprocessing/labextraction`.

O prompt tem uma barreira para retornar `tests: []` em laudos de imagem, atestados,
pedidos, receitas e prescrições.

O adaptador trata:
- entrada vazia;
- erro do cliente;
- resposta sem candidato;
- resposta bloqueada;
- resposta truncada por limite de tokens;
- texto vazio;
- JSON inválido;
- JSON fora do schema;
- resposta válida sem itens, devolvida sem avaliação de qualidade pelo adaptador.

O serviço `documentprocessing/extraction` prepara o texto semântico, preserva o
texto original e aplica normalização, avisos e estados (`needs_review`, `partial`
ou `succeeded`). O adaptador valida a resposta e informa fornecedor/modelo.

Não registrar credenciais, texto clínico ou respostas completas nos logs comuns.

## Testes

`NewClientWithGenerator` recebe a interface `ContentGenerator`. Os testes injetam
uma implementação simulada, sem credenciais reais nem acesso a rede.

```powershell
go test ./internal/config ./internal/infrastructure/gemini -v
```
