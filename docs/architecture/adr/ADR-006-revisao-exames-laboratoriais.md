<!-- docs/architecture/adr/ADR-006-revisao-exames-laboratoriais.md -->
# ADR-006 — Extração compartilhada e confirmação de exames

Status: implementada. Substitui as decisões de orquestração/persistência anteriores da ADR-005; preserva o extrator semântico e seu contrato.

## Fluxos

- Área de trabalho: PDF com texto selecionável → leitura local → Gemini → resumo copiável. Não grava banco nem armazenamento permanente; o arquivo temporário é removido ao terminar, inclusive em falhas.
- Paciente: verificar acesso → extrair → salvar PDF no GCS e rascunho no Postgres → conferir → confirmar no histórico. Não grava resultados clínicos antes da confirmação.
- Terminal: mantém `extract-lab-summary -input texto.txt [-output resumo.txt]` e utiliza o mesmo serviço de extração de texto e formatação.

Limite de 10 MiB por PDF. Processamento síncrono, sem fila e sem OCR remoto nos endpoints web. O Document AI continua disponível no comando independente de extração de texto. Falhas anteriores à criação do rascunho exigem novo envio.

## Responsabilidades

`documentprocessing/extraction` não depende de paciente, storage nem banco. Retorna `status`, `warnings`, `report` e `summary_text`. Normaliza datas e dados antes da conferência, sem converter unidades ou valores; entradas sem identificação são omitidas com aviso. Resultado utilizável contém ao menos um exame e parâmetro identificados com valor, inclusive qualitativo.

`documentprocessing` cuida do documento, da fotografia da extração e da exclusão. `patient/exam/laboratory` mantém os modelos e a leitura do histórico clínico. `application/usecase/labdocumentconfirmation` converte a fotografia conferida em exame e solicita a gravação transacional.

Não há classificação automática nem processadores de imagem neste fluxo. A seleção explícita de uma funcionalidade laboratorial define o tipo. PDFs de outros tipos não geram rascunhos sem resultados laboratoriais utilizáveis.

## Persistência

- `exam_documents.status` representa processamento; `review_status` representa `pending`, `confirmed` ou `deleting`.
- `exam_document_extractions` tem relação 1:1 com o documento. Guarda JSON versionado com resultado público e metadados internos que o JSON público omite, incluindo texto e avisos por item. Um trigger impede atualização da fotografia.
- A criação do documento e da fotografia usa uma transação. Se falhar, há tentativa de compensação do upload; falha de compensação permanece na cadeia interna de erro para investigação.
- A confirmação bloqueia o documento e grava laudo, resultados, vínculo, fingerprint e autor/data da confirmação na mesma transação. Repetições retornam o mesmo exame. O Gemini não é chamado na confirmação.
- A exclusão marca `deleting`, remove o objeto e depois o registro. Falhas de storage preservam o rascunho para nova tentativa. Objetos já ausentes são tratados como removidos. Documentos confirmados e anteriores não podem ser excluídos por essa operação.
- Documentos anteriores permanecem com `review_status` nulo. Não se atribui confirmação humana retroativa.
- Não há novas gravações em `exam_document_texts`; a consulta dos textos antigos permanece disponível.
- As tabelas documentais têm RLS e acesso direto de `anon`/`authenticated` revogado. O backend verifica acesso ao paciente em cada operação.

## Contratos e compatibilidade

O upload `POST /patients/{patientId}/exam-documents` retorna um rascunho (`201`), nunca um exame automaticamente registrado. O campo `collection_date` foi removido; cada resultado preserva a data encontrada no PDF.

- `GET /exam-documents/{documentId}/extraction`: fotografia para conferência.
- `POST /exam-documents/{documentId}/confirmation`: sem dados clínicos no corpo; retorna o exame confirmado (`200`).
- `DELETE /exam-documents/{documentId}`: descarta rascunho e arquivo (`204`). Uma repetição após exclusão concluída retorna `404`, que o web trata como conclusão.
- Consultas de documentos, PDFs assinados e laudos continuam disponíveis.
- Aliases `/patients/{patientId}/exames` foram removidos.

O web mostra PDF, resumo, dados completos e avisos; não permite edição. Exige conferência explícita antes de confirmar e aceita resultados parciais. Pendências podem ser retomadas em outra sessão. Rascunhos não expiram automaticamente.

**Quebra para o mobile:** a aplicação antiga deixa de registrar resultados ao enviar PDF. Adaptar o mobile para buscar a extração e chamar a confirmação em uma entrega posterior. Não restaurar a gravação automática como compatibilidade.

## Implantação e verificação

1. Aplicar `supabase/migrations/20260930211213_lab_document_review.sql` no ambiente de destino pelo processo de migrations do projeto.
2. Liberar API e web de forma coordenada. Publicar o OpenAPI identificado pelo SHA da API e gerar o consumidor web desse artefato.
3. Verificar extração temporária, criação/retomada de rascunho, confirmação repetida, exclusão e leitura de exames anteriores.

O espelho sqlc da migration serve ao fluxo SQL local; não aplicar ambos os arquivos no mesmo banco. A migration é aditiva e não remove histórico. O rollback da aplicação não deve reativar uploads automáticos enquanto existirem clientes no novo fluxo.

Validações automatizadas:

```text
go test ./...
go test -tags integration ./internal/features/documentprocessing/postgres ./internal/features/patient/exam/laboratory/postgres
go tool sqlc compile -f internal/infrastructure/persistence/postgres/sqlc/sqlc.yaml
go run ./cmd/openapi-export -output artifacts/openapi.json -version <API_SHA>
bun run openapi:generate
bun run test
bun run lint
bun run build
```

Integrações exigem `LABS_TEST_DATABASE_URL` apontando para Postgres local isolado. Os testes criam schemas temporários; o teste da migration cria, se necessário, os papéis locais `anon`, `authenticated` e `service_role`.

Monitorar erros de extração, confirmação e compensação pelo logging centralizado. Não registrar laudos, resultados ou conteúdo de PDFs nos logs.
