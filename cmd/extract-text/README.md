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

Nesse modo, o metodo selecionado e exibido em `stderr`; o texto continua em
`stdout`, permitindo redireciona-lo para outro comando ou arquivo.

PDF nativo usa `pdftotext -raw`. Imagem `.jpg`, `.jpeg` ou `.png` usa `tesseract`.
Para fotos, o comando testa rotacoes de 0, 90, 180 e 270 graus, depois cria versoes
temporarias em tons de cinza com contraste alto e binarizacao. A melhor rotacao e
avaliada primeiro pelo modo padrao `psm 3`, e depois pelos modos `psm 4`, `psm 6`
e `psm 11`; a leitura com contraste alto e a binarizada usam `psm 6`. As fotos sao
limitadas a 2200 px no maior lado durante o OCR. Apenas o texto vencedor e salvo,
e as variacoes de imagem sao removidas ao fim da extracao.
PDF escaneado sem camada de texto ainda nao e convertido para imagem nesta etapa.
