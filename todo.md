# Extração laboratorial compartilhada e confirmação de exames

## 1. Objetivo e decisões

Organizar API e web em dois fluxos que compartilham a mesma extração:

- **Área de trabalho:** PDF → extração → resumo copiável, sem persistência em banco ou armazenamento de arquivos.
- **Paciente:** PDF → extração → rascunho persistente → conferência → confirmação no histórico.

Nesta entrega:

- Apenas exames laboratoriais em PDF com texto selecionável, até 10 MB.
- Processamento síncrono; falhas antes de salvar o rascunho exigem reenviar o PDF.
- Conferência sem edição dos resultados.
- Extrações parciais podem ser confirmadas, com avisos visíveis.
- Datas vêm do PDF; remover a substituição manual por uma data única.
- Descartar um rascunho exclui seu PDF e sua extração.
- Classificação automática, exames de imagem, OCR remoto e adaptação do mobile ficam para depois.

## 2. Organização e responsabilidades

**Extração compartilhada**

Concentrar em `internal/features/documentprocessing/extraction` o fluxo de leitura local, extração semântica, normalização, avaliação do resultado e geração do resumo.

- Reutilizar `pdftotext` e o extrator Gemini existentes.
- Retornar um resultado comum: `status`, `warnings`, `report` e `summary_text`.
- Não depender de paciente, repositórios ou armazenamento permanente.
- Gerar o resumo deterministicamente a partir dos dados estruturados: paciente, laboratório, datas, exames, parâmetros, valores e unidades.
- Preservar valores textuais, comparadores, unidades e datas distintas por exame; campos ausentes continuam ausentes.
- Distinguir PDF ilegível de falhas técnicas, indisponibilidade do extrator e timeout.

**Documentos e rascunhos**

`documentprocessing` permanece responsável pelo arquivo original, extração armazenada e ciclo de vida do rascunho. O upload do paciente verifica acesso, extrai o PDF e somente então salva arquivo e rascunho.

Uma extração utilizável deve conter ao menos um exame e parâmetro identificados com valor não vazio, incluindo resultados qualitativos. Sem resultados utilizáveis, o fluxo do paciente retorna erro de validação e não cria rascunho.

**Histórico laboratorial**

`internal/features/patient/exam/laboratory` continua responsável pelos exames clínicos confirmados.

Um caso de uso em `internal/application/usecase/labdocumentconfirmation` coordena a confirmação: lê a extração armazenada, aplica as regras laboratoriais e grava o exame. A confirmação não executa novamente o Gemini nem recebe resultados editáveis do navegador.

**Terminal e limpeza**

- Adaptar o comando de resumo para consumir a mesma extração de texto e formatação, preservando seus argumentos atuais.
- Remover do fluxo ativo o classificador heurístico, o registro genérico de processadores e a persistência automática.
- Remover código que ficar sem consumidores e atualizar os documentos de arquitetura afetados.
- Desconectar Document AI da inicialização desses fluxos; preservar ferramentas independentes que ainda o utilizem.

## 3. Persistência e contratos da API

Separar a qualidade da extração da decisão do usuário:

- O status de processamento descreve a extração.
- Um novo `review_status` indica `pending`, `confirmed` ou `deleting`.
- Documentos anteriores mantêm `review_status` ausente, sem inventar uma confirmação humana.

Adicionar ao documento uma fotografia imutável da extração, com versão de formato, dados estruturados, avisos, resumo e metadados internos. Usar serialização própria para persistência: os DTOs atuais omitem metadados importantes com `json:"-"`.

| Operação                                         | Comportamento                                                                                       |
| ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `POST /lab-extractions`                          | Retorna extração e `summary_text`; remove o arquivo temporário ao terminar.                         |
| `POST /patients/{patientId}/exam-documents`      | Salva um rascunho laboratorial após extração utilizável; retorna `201` com `review_status=pending`. |
| `GET /patients/{patientId}/exam-documents`       | Lista documentos com estado de revisão e vínculo ao exame confirmado, sem carregar toda a extração. |
| `GET /exam-documents/{documentId}/extraction`    | Retorna a extração armazenada para conferência.                                                     |
| `POST /exam-documents/{documentId}/confirmation` | Confirma o rascunho e retorna o exame persistido.                                                   |
| `DELETE /exam-documents/{documentId}`            | Exclui somente rascunhos; documentos confirmados ou anteriores retornam conflito.                   |

Manter os endpoints de detalhe, arquivo assinado e consulta de exames. Remover os aliases antigos em `/exames`, evitando caminhos que continuem salvando automaticamente.

**Garantias de gravação**

- Criar documento e fotografia da extração na mesma transação; compensar o upload se essa gravação falhar.
- Confirmar em uma única transação: bloquear o rascunho, gravar laudo/resultados/metadados e registrar autor e horário da confirmação.
- Repetir a confirmação retorna o mesmo exame. Preservar a proteção existente contra exames duplicados.
- Para excluir, marcar `deleting`, bloquear confirmação, remover o arquivo e finalizar a remoção no banco. Falhas mantêm a referência necessária para repetir a exclusão; arquivo já removido conta como sucesso.
- Não manter transações abertas durante chamadas ao Gemini ou ao armazenamento.
- Verificar acesso ao paciente em todas as operações, inclusive confirmação, exclusão e obtenção do PDF.
- Usar `AppError`, respostas Huma e a política central de logs; não registrar conteúdo dos laudos.

## 4. Experiência web e migração

**Área de trabalho:** manter upload e resumo lado a lado, usar `summary_text` da API e oferecer cópia. Estado permanece apenas na sessão da página, sem cache persistente.

**Paciente:** apresentar pendências e histórico confirmado separadamente. Abrir o rascunho com PDF, resumo, dados estruturados completos e avisos; oferecer **Confirmar exame** e **Descartar rascunho**. A conferência mostra também referências e demais campos que serão persistidos, mesmo quando não aparecem no resumo.

- Permitir retomar rascunhos após recarregar a página.
- Exibir explicitamente o paciente selecionado e a identificação extraída.
- Bloquear ações repetidas enquanto a requisição está em andamento.
- Limpar o estado ao trocar de paciente ou encerrar a sessão.
- Atualizar o histórico somente após confirmação.
- Corrigir os rótulos de status: a API atual retorna valores em minúsculas.

A migração será aditiva, preservando exames e documentos anteriores. Novos fluxos deixam de produzir cópias em `exam_document_texts`; registros existentes continuam consultáveis.

Atualizar migrations, consultas e schema do sqlc, regenerar o código e gerar o cliente web pelo OpenAPI do Huma. Proteger as tabelas envolvidas contra acesso direto não autorizado pela Data API.

API e web estão em repositórios separados: a entrega terá alterações coordenadas, com implantação na ordem **migração → API → web**. Registrar a mudança de contrato para a futura adaptação do mobile.

## 5. Validação e premissas

**Testes de aceitação**

- O mesmo PDF produz os mesmos dados e resumo nos dois fluxos.
- A extração temporária não chama persistência e remove temporários também em falhas.
- Upload do paciente cria rascunho, sem inserir resultados no histórico.
- Reabrir rascunho preserva exatamente a extração conferida.
- Extração parcial mantém avisos; ausência de resultados utilizáveis impede confirmação.
- Dupla confirmação, concorrência e repetição após perda de resposta não duplicam exames.
- Falhas durante confirmação revertem toda a transação.
- Exclusão não remove exames confirmados e pode ser repetida após falha de armazenamento.
- Usuários sem acesso ao paciente não consultam nem alteram seus documentos.
- PDFs escaneados, arquivos inválidos e uploads acima do limite recebem erros adequados.
- Documentos e exames anteriores continuam legíveis.

Executar testes Go, integrações transacionais em Postgres local isolado, validação do sqlc/OpenAPI e testes, lint e build do web. Os testes existentes de processamento, laboratório e comando de resumo passaram durante o levantamento.

**Premissas:** preservar o GCS encontrado no código como armazenamento de arquivos; manter Supabase/Postgres para o banco. Rascunhos permanecem até confirmação ou exclusão explícita. Não adicionar fila, expiração automática, edição clínica ou interpretação médica neste PR.
