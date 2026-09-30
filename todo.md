Sim, mas bem menor enquanto houver endpoints/middlewares Gin legados.

Para as rotas Huma, mantenha apenas a tradução do `AppError`:

- `StatusFromCode`: converte o código interno para status HTTP.
- Uma conversão `AppError → huma.StatusError` que preserve `code` e `violations`, caso esses campos façam parte do contrato público.
- A política de logging/observabilidade, idealmente aplicada também aos erros retornados pelo Huma.

Hoje há uma inconsistência: [`writeHumaError`](/C:/Users/gabri/Dev/sonnda/sonnda-api/internal/api/huma.go:133) monta um `Problem`, mas passa apenas `problem.Detail` para `huma.WriteErr`; portanto `code`, `violations`, `traceId` e `instance` são descartados. Já handlers Huma usam `huma.NewError(...)`, cujo modelo padrão é RFC 9457, mas também não inclui o `code` Sonnda.

O que ainda precisa ficar para Gin é `ErrorResponder`, `writeProblem` e os helpers de log, pois existem três rotas legadas em [`routes.go`](/C:/Users/gabri/Dev/sonnda/sonnda-api/internal/api/routes.go:59), além de middlewares Gin.

Quando essas rotas forem migradas para Huma, eu removeria `ErrorResponder` e a escrita manual de JSON. O diretório poderia virar algo como `internal/api/problem`, contendo só:

- `StatusFromCode`
- adaptador `AppError → huma.StatusError`/modelo de problema Huma
- mapeamento de violações para `huma.ErrorDetail`, se necessário

Ou seja: Huma substitui o “writer” do presenter; não substitui automaticamente a tradução do contrato de erro da aplicação.