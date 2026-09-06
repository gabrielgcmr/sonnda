<!-- cmd/extract-text/README.md -->
# extract-text

Comando local para gerar `.txt` a partir de PDFs e imagens usando o mesmo extrator
local do backend.

## Uso

Converter todos os arquivos brutos de `samples/exam`:

```powershell
go run ./cmd/extract-text -input samples/exam -output samples/text
```

Por padrao, o comando salva o texto bruto mesmo quando ele nao passa no filtro de
qualidade do runtime. Para aplicar o mesmo filtro da API:

```powershell
go run ./cmd/extract-text -input samples/exam -output samples/text -require-usable
```

Converter um arquivo e imprimir no terminal:

```powershell
go run ./cmd/extract-text -input "samples/exam/exemplo.pdf"
```

PDF nativo usa `pdftotext -raw`. Imagem `.jpg`, `.jpeg` ou `.png` usa `tesseract`.
PDF escaneado sem camada de texto ainda nao e convertido para imagem nesta etapa.
