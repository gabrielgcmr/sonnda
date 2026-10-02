<!-- internal/features/documentprocessing/extraction/README.md -->

# Coordenação da extração

`Service` coordena a extração de um PDF ou de um texto já lido. Ele recebe o
texto bruto, preserva esse conteúdo para auditoria, normaliza o texto usado na
interpretação semântica, avalia a qualidade do resultado e monta o resumo que a
API devolve.

O serviço não conhece paciente, banco, armazenamento permanente ou HTTP. Os
consumidores atuais são:

- o fluxo temporário de `POST /lab-extractions`, que devolve o resultado sem
  persistir dados;
- o fluxo de rascunhos de documentos, que guarda um snapshot da extração e só
  cria o histórico clínico depois da confirmação.

O resultado distingue sucesso completo de resultado parcial e carrega avisos
para a conferência. Falhas de leitura e falhas técnicas continuam sendo
tratadas pelos contratos de `textextraction` e `apperr`; o serviço não inventa
um texto quando o PDF é ilegível.

O resumo é formatado aqui para que os dois fluxos apresentem o mesmo cabeçalho,
resultados e avisos. A confirmação posterior transforma o snapshot em
`lab_reports`, `lab_panels` e `observations`; ela não executa a extração outra
vez.
