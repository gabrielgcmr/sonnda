<!-- internal/features/documentprocessing/README.md -->
# Document Processing

Feature responsável pelo processamento de documentos clínicos e laudos laboratoriais, englobando leitura textual, extração estruturada semântica, gerenciamento de rascunhos e fotografia persistida (snapshot).

## Estrutura de Pacotes

```text
internal/features/documentprocessing/
├── queries.go              # Consultas de documentos e textos extraídos
├── drafts.go               # Coordenação de criação, conferência e exclusão de rascunhos
├── snapshot.go             # Codificação e decodificação do snapshot persistido
├── dto.go                  # DTOs públicos de documentos
├── repository.go           # Contratos de persistência de documentos
├── textextraction/         # Contrato de leitura, qualidade e normalização de texto
├── labextraction/          # Contrato, tipos e schema da extração estruturada
├── extraction/             # Coordenação, normalização dos dados, avaliação e resumo
├── http/                   # Handlers Huma (extração temporária e rascunhos)
└── postgres/               # Adaptadores PostgreSQL para documentos e rascunhos
```

## Direção das Dependências

1. **`textextraction` e `labextraction` (Folhas)**:
   - Definem interfaces (`Extractor`, `LabReportTextExtractor`), schemas e DTOs puros.
   - Não dependem da raiz da feature, de HTTP, banco de dados ou SDKs de terceiros.
2. **`extraction` (Coordenação de extração)**:
   - Compõe `textextraction.Extractor` e `labextraction.LabReportTextExtractor`.
   - Executa normalização de entrada semântica, avaliação de resultados e geração de resumo.
   - Não depende de HTTP, banco de dados, storage nem da raiz de `documentprocessing`.
3. **Raiz de `documentprocessing`**:
   - `queries.go` expõe a interface pública interna `Service` para consultas de documentos.
   - `drafts.go` coordena a extração via `PDFExtractor`, upload no storage e persistência atômica do rascunho com snapshot.
   - `snapshot.go` (`EncodeExtractionSnapshot` / `DecodeExtractionSnapshot`) serializa `extraction.Result` preservando metadados privados (texto bruto, status, confiança e avisos por item). Fica na raiz para evitar dependências circulares com `extraction`.
4. **Infraestrutura e Casos de Uso Externos**:
   - Adapters concretos de texto e Gemini vivem em `internal/infrastructure/textextraction` e `internal/infrastructure/gemini`.
   - A conversão do snapshot em histórico clínico é responsabilidade do caso de uso `internal/application/usecase/labdocumentconfirmation`.

## Consumidores da Extração

O pipeline de extração atende a dois fluxos:

1. **Extração temporária (`StandaloneLabExtractionHandler`)**:
   - Rota `POST /lab-extractions`.
   - Utilizada para conferência imediata em área de trabalho.
   - O PDF temporário é descartado após o processamento, inclusive em falhas. Não grava em banco nem em storage persistente.
2. **Criação de rascunho (`Drafts`)**:
   - Rota `POST /patients/{patientId}/exam-documents`.
   - Grava o PDF original no storage, persiste o snapshot da extração e gera um documento com status `pending`.
   - Permite conferência posterior e confirmação transacional no histórico do paciente.

> **Nota:** Utilitários de extração via terminal/CLI foram descontinuados e removidos para restringir a execução exclusivamente às rotas autenticadas da API.
