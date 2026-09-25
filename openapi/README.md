<!-- openapi/README.md -->
# Contrato OpenAPI

O ponto de entrada continua em `../openapi.yaml`. Edite esse arquivo e os módulos;
a versão servida pela API é um bundle sem referências a arquivos externos.

Account é o primeiro contexto extraído:

- `paths/account.yaml`: operações de perfil em `/v1/me`.
- `paths/patient-access.yaml`: listagem de `/v1/me/patients`.
- `components/schemas/account.yaml`: `CreateUserRequest`, `UpdateUserRequest` e `User`.

Os demais contratos e componentes compartilhados continuam no arquivo principal.
Os `$ref` são relativos ao arquivo que os contém. Assim, as rotas de account
referenciam os componentes compartilhados por `../../openapi.yaml#/components/...`.
Não alteramos campos, respostas ou regras de validação nesta extração.

## Validar e gerar

```sh
go run ./cmd/openapi-validate -file openapi.yaml
go run ./cmd/openapi-bundle -input openapi.yaml -output bin/openapi.yaml
go generate ./internal/api/openapi
go tool oapi-codegen -generate types,gin -package openapi -o internal/api/openapi/generated/oapi.gen.go bin/openapi.yaml
```

Com Make, use `make openapi-validate`, `make openapi-bundle`,
`make openapi-embed` e `make oapi-codegen`. O último gera o bundle antes dos DTOs.
`make generate` também executa SQLC.

A validação e o bundling usam `kin-openapi`, já presente no projeto. Referências
ausentes ou inválidas interrompem a geração. Os nomes públicos dos schemas são
preservados no bundle para manter estáveis os tipos gerados em Go.

Os tipos de transporte gerados continuam separados dos modelos de domínio.
Mudanças futuras de contrato devem começar nos YAMLs e atualizar a geração,
os mapeamentos HTTP e os testes correspondentes.
