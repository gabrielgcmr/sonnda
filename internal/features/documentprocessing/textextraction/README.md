<!-- internal/features/documentprocessing/textextraction/README.md -->
# Extração textual

Este pacote define o contrato puro para leitura de documentos antes da extração laboratorial estruturada.

`Extractor` recebe a origem local ou remota, o tipo MIME e o nome original. O adaptador retorna o texto bruto, uma versão normalizada para uso semântico e o método utilizado. A implementação concreta fica em `internal/infrastructure/textextraction` e não deve conhecer HTTP, banco, paciente ou Gemini.

`documentprocessing/extraction` consome este contrato, preserva `Text` no snapshot e envia somente `NormalizedText` ao extrator laboratorial. A qualidade mínima do texto é avaliada por `IsUsableText`; um PDF sem texto selecionável é uma falha de entrada, enquanto indisponibilidade do leitor é uma falha técnica.

O fluxo atual usa leitura local síncrona com `pdftotext` para PDFs. OCR remoto e fallback automático permanecem fora deste fluxo.
