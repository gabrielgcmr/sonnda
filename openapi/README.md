<!-- openapi/README.md -->
# Contrato OpenAPI

O ponto de entrada continua em `../openapi.yaml`. Esse arquivo e os módulos são
a fonte de verdade editável. `../dist/openapi.yaml` é o bundle gerado e é o único
YAML consumido pelo embed, pelo `oapi-codegen` e por futuros geradores de clientes.

Os contratos já extraídos por contexto são:

- `paths/account.yaml`: operações de perfil em `/v1/me`.
- `paths/patient-access.yaml`: listagem de `/v1/me/patients`.
- `paths/patient.yaml`: criação, listagem e consulta do perfil do paciente.
- `components/schemas/account.yaml`: `CreateUserRequest`, `UpdateUserRequest` e `User`.
- `components/schemas/patient-access.yaml`: respostas da listagem de acesso.
- `components/schemas/patient.yaml`: requests e responses do perfil do paciente.

Os demais contratos e componentes compartilhados continuam no arquivo principal.
Os `$ref` são relativos ao arquivo que os contém. Assim, os módulos de paths
referenciam os componentes compartilhados por `../../openapi.yaml#/components/...`.

## Convenções de operações

- Toda operação deve declarar um `operationId` único, estável e em `lowerCamelCase`.
- Parâmetros de caminho devem descrever o recurso, como `{patientId}` e `{problemId}`; não use `{id}`.
- Parâmetros reutilizados devem ser declarados em `components/parameters`, como `PatientId`.
- Renomear um `operationId` é uma mudança de contrato porque altera os métodos dos clientes gerados.

## Validar e gerar

```sh
go run ./cmd/openapi-validate -file openapi.yaml
go run ./cmd/openapi-bundle -input openapi.yaml -output dist/openapi.yaml
go run ./cmd/openapi-validate -file dist/openapi.yaml
go generate ./internal/openapispec
go tool oapi-codegen -generate types,gin -package openapi -o internal/generated/openapi/oapi.gen.go dist/openapi.yaml
```

Com Make, `make openapi-generate` executa toda a sequência: valida a fonte,
gera e valida o bundle, atualiza o spec embutido e gera os tipos Go.
Os targets individuais continuam disponíveis. `make generate` também executa SQLC.

A validação e o bundling usam `kin-openapi`, já presente no projeto. Referências
ausentes ou inválidas interrompem a geração. Os nomes públicos dos schemas são
preservados no bundle para manter estáveis os tipos gerados em Go.

O bundle fica em `dist/openapi.yaml`, os tipos Go em
`internal/generated/openapi/oapi.gen.go` e os bytes servidos pela API em
`internal/openapispec/spec.gen.go`. Os dois arquivos Go são gerados exclusivamente
a partir do bundle.

Os tipos de transporte gerados continuam separados dos modelos de domínio.
Mudanças futuras de contrato devem começar nos YAMLs e atualizar a geração,
os mapeamentos HTTP e os testes correspondentes.
