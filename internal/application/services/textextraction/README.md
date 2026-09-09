<!-- internal/application/services/textextraction/README.md -->
# Leitura de documentos com segunda tentativa

O upload salva o arquivo antes da extracao. O extrator local tenta ler o texto
usando pdftotext ou Tesseract. Em caso de erro, timeout local, resposta vazia ou
texto que nao passa em `IsUsableText`, o servico tenta Document AI com a URI GCS
do mesmo arquivo. Um resultado local aproveitavel evita essa chamada adicional.

O runtime reutiliza o cliente e o processador `GCP_EXTRACT_LABS_PROCESSOR_ID`
ja configurados. O adaptador de texto utiliza `Document.text`, sem exigir as
entidades laboratoriais. A resposta de OCR faz parte da saida dos processadores:
[contrato do Document AI](https://docs.cloud.google.com/document-ai/docs/handle-response).

`OCR_TIMEOUT` limita a etapa local (padrao 60s). `OCR_FALLBACK_TIMEOUT` limita a
tentativa remota (padrao 60s). A segunda tentativa usa o contexto da requisicao,
nao o contexto expirado da primeira; cancelamento da requisicao interrompe o fluxo.
Nao ha repeticao automatica de chamadas remotas.

O texto recuperado segue para a classificacao existente com metodo
`document_ai_ocr`. Se classificado como laboratorial, o fluxo atual ainda chama
Document AI para extrair resultados estruturados: nesse caso sao duas chamadas
remotas, com o custo correspondente. O texto bruto continua separado dos valores
estruturados; recuperar OCR nao garante extracao correta dos resultados.

Se nenhuma tentativa funcionar, um AppError conserva as causas internas e uma
mensagem segura informa a necessidade de revisao. Os logs
`exam_ocr_attempt_failed` e `exam_ocr_attempt_succeeded` registram provedor e
duracao no contexto da requisicao, sem registrar o texto do laudo.

Este fluxo ainda e sincrono. O tempo da requisicao pode incluir as duas leituras
e a extracao laboratorial; limites de proxy precisam comportar esse tempo.
Uma fila persistente com worker e reprocessamento e uma evolucao separada.
Nao ha migracao nem reprocessamento automatico dos documentos antigos.
