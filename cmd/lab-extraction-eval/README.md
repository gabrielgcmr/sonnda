<!-- cmd/lab-extraction-eval/README.md -->
# lab-extraction-eval

Comando local para avaliar o extrator laboratorial Gemini com texto ja extraido.
Ele nao altera banco, nao chama rotas HTTP e nao envia arquivos PDF/imagem.

Para gerar `.txt` a partir dos arquivos brutos de `samples/exam`, use primeiro:

```powershell
go run ./cmd/extract-text -input samples/exam -output samples/text
```

## Uso

Configure `GEMINI_API_KEY` no ambiente ou no `.env` local:

```powershell
$env:GEMINI_API_KEY="sua-chave"
```

Execute com um arquivo `.txt`:

```powershell
go run ./cmd/lab-extraction-eval -input samples/labs/glicose.txt
```

Antes de chamar o Gemini, o comando usa o classificador local de documentos. Se o
texto parecer laudo de imagem, atestado, pedido ou outro documento nao laboratorial,
a chamada e pulada para evitar custo.

Para testar o prompt manualmente mesmo assim:

```powershell
go run ./cmd/lab-extraction-eval -input samples/text/LuzUSG.txt -force-lab
```

Para comparar com um JSON esperado:

```powershell
go run ./cmd/lab-extraction-eval `
  -input samples/labs/glicose.txt `
  -expected samples/labs/glicose.expected.json
```

Quando o gate local pula a chamada, `-expected` nao e comparado porque nao ha
resposta do extrator laboratorial.

A saida principal vai para `stdout` como JSON no contrato de `ExtractedLabReport`.
Mensagens de erro ou comparacao vao para `stderr`.

Codigos de saida:

- `0`: extracao executada; quando `-expected` for informado, JSON igual ao esperado;
- `1`: erro de configuracao, leitura, chamada do provedor ou validacao;
- `2`: argumento obrigatorio ausente ou JSON gerado diferente do esperado.

Use somente textos sinteticos ou anonimizados ate a politica de dados reais estar
revisada. O comando imprime o resultado extraido no terminal.
