# Refatoração de `documentprocessing` em etapas

## Objetivo e estrutura

Concentrar os contratos de processamento na feature, explicitar a responsabilidade de cada etapa e preservar o comportamento atual. Executar em seis commits sequenciais, mantendo o projeto compilando e os testes relevantes passando em cada um.

A organização final terá:

- `documentprocessing/textextraction`: contrato de leitura, qualidade e normalização de texto.
- `documentprocessing/labextraction`: contrato, tipos e schema da extração estruturada.
- `documentprocessing/extraction`: coordenação, normalização dos dados, avaliação e resumo.
- Raiz de `documentprocessing`: documentos, consultas, rascunhos e snapshot persistido.
- Infraestrutura: implementações de leitura e integração Gemini.
- `labdocumentconfirmation`: conversão do snapshot em histórico clínico.

## Etapas de implementação

### 1. Fixar a referência de comportamento

Complementar os testes existentes com casos de caracterização da extração completa: sucesso, resultado parcial, ausência de resultados, campos sem identificação, datas inválidas e valores qualitativos.

Registrar respostas esperadas, incluindo códigos e ordem dos avisos, estados, resumo e metadados internos. Criar uma fixture de snapshot versão 1 para verificar sua leitura depois da refatoração. Exportar o OpenAPI atual para um arquivo temporário de comparação.

**Concluída quando:** os testes representam o comportamento vigente e passam antes das mudanças estruturais.

### 2. Reorganizar contratos e adaptadores

Mover os pacotes globais de extração textual e laboratorial para dentro de `documentprocessing`, levando testes, fixtures, documentação e schema embarcado. Manter nomes dos tipos, interfaces, campos e versões.

Mover Gemini de `infrastructure/persistence/gemini` para `infrastructure/gemini`. Atualizar imports da API, bootstrap, comandos, Document AI e testes no mesmo commit, sem deixar aliases de compatibilidade.

Os pacotes de contratos não dependerão da raiz da feature, de HTTP, banco ou SDKs.

**Concluída quando:** todos os consumidores compilam e os testes dos pacotes movidos passam.

### 3. Centralizar a composição do fluxo

Alterar o construtor do handler de extração temporária para receber uma interface com `ExtractPDF`, em vez dos extratores de texto e laboratório.

O bootstrap construirá o serviço e fornecerá a mesma instância à extração temporária e aos rascunhos. O terminal continuará compondo suas dependências e utilizando `ExtractText`, sem exigir um leitor de PDF.

Manter `Drafts` como coordenador de extração, armazenamento e criação do documento. A confirmação continuará sendo um caso de uso separado.

**Concluída quando:** handlers não constroem serviços de extração e os testes comprovam que os dois fluxos HTTP usam a dependência injetada.

### 4. Consolidar as regras de extração

Separar, dentro do pacote `extraction`, funções para normalização, avaliação de resultados e formatação. Manter a sequência explícita no serviço, sem criar um mecanismo genérico de etapas.

Gemini ficará responsável por preparar a requisição, chamar o fornecedor, validar término/JSON/schema e informar fornecedor/modelo. Transferir suas decisões sobre ausência de resultados e estados dos itens para a aplicação.

Executar a normalização de entrada semântica na aplicação, preservando o texto original. Consolidar as chamadas redundantes de normalização dos dados estruturados.

Preservar as regras atuais: presença de estrutura e presença de resultado utilizável continuam sendo verificações distintas. A mudança de localização não alterará avisos, estados ou critérios de aceitação.

**Concluída quando:** os testes de caracterização continuam passando e as regras de qualidade podem ser testadas sem Gemini.

### 5. Separar snapshot e remover componentes sem uso

Mover a codificação do snapshot para a raiz de `documentprocessing`, com funções `EncodeExtractionSnapshot` e `DecodeExtractionSnapshot`. Atualizar rascunhos e testes de persistência.

Preservar integralmente a versão 1, seus campos privados e validações. Separar os testes de processamento dos testes de serialização para evitar dependências circulares.

Remover o classificador heurístico, o fallback textual sem consumidores de produção e o helper HTTP `isNilExtractor`, junto dos testes exclusivos desses componentes. Preservar o comando de leitura com Document AI e as funcionalidades efetivamente usadas.

**Concluída quando:** snapshots anteriores são recuperados sem perda e a busca de referências confirma a remoção dos componentes selecionados.

### 6. Documentar e validar a entrega

Atualizar as instruções do repositório, documentação arquitetural, READMEs dos pacotes e ADR-006. Registrar a reorganização como complemento histórico da ADR-005.

Renomear `service_impl.go` para `queries.go`, mantendo a interface pública interna de consultas. Documentar a direção das dependências e os três consumidores da extração: temporário, rascunho e terminal.

**Concluída quando:** as verificações abaixo passam e o diff contém somente a refatoração planejada.

## Validação

- Executar testes dos pacotes afetados em cada etapa e `go test ./...` ao concluir.
- Verificar PDF ilegível, falha do fornecedor, cancelamento, ausência de resultado e limpeza do arquivo temporário.
- Verificar preservação de texto original, valores, unidades, datas, avisos, resumo e metadados no snapshot.
- Executar os testes de integração existentes em Postgres local isolado: criação sem histórico clínico, confirmação atômica e idempotente, rollback e retomada da exclusão.
- Confirmar que a confirmação utiliza o snapshot e não repete a extração.
- Comparar o OpenAPI antes/depois com a mesma versão de exportação; não aceitar mudanças no contrato HTTP.

Os testes unitários direcionados executados durante o planejamento passaram. Os testes de integração ainda não foram executados.

## Premissas e limites

- Reorganização completa neste PR, em commits separados.
- Permanecem o processamento síncrono, os endpoints, as respostas e a confirmação explícita.
- Não haverá migrations, mudanças no schema de extração, fila, novo fornecedor ou alterações no web/mobile.
- Preservar o tratamento centralizado de erros e logging.
- Preservar a exclusão preexistente de `docs/dev/setup.md`, fora desta refatoração.
