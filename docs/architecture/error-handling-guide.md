# Guia de Tratamento de Erros — Sonnda API

**Última atualização:** 31 de dezembro de 2025  
**Referências:** [ADR-006](adr/ADR-006-apperr-kind-based-error-handling.md), `internal/app/apperr`, `internal/http/api/handlers/common`

---

## Visão Geral

A Sonnda API implementa um modelo de tratamento de erros baseado em **categorias semânticas** (Kinds), desacoplando o domínio/aplicação da borda HTTP.

**Princípios:**
- Domínio e Services não conhecem HTTP
- Erros carregam significado semântico (`Kind`), não status HTTP
- Camada HTTP traduz `Kind → Status HTTP` em ponto único
- Logging estruturado e consistente para todos os erros

---

## Estrutura de Erro (`apperr.Error`)

Todos os erros da aplicação seguem a estrutura:

```go
type Error struct {
	Kind    Kind   // Categoria semântica (ex: not_found, invalid_input)
	Code    string // Identificador estável (ex: "patient_not_found")
	Message string // Descrição humana (opcional)
	Err     error  // Causa original (wrap)
}
```

### Kinds Disponíveis

| Kind | Significado | Uso Típico |
|------|-------------|------------|
| `KindInvalidInput` | Dados de entrada inválidos | Validação de campos, parsing |
| `KindNotFound` | Recurso não encontrado | Entidade não existe no banco |
| `KindConflict` | Conflito de estado | CPF duplicado, recurso já existe |
| `KindUnauthorized` | Não autenticado | Token ausente/inválido |
| `KindForbidden` | Sem permissão | Policy negou acesso |
| `KindExternal` | Erro em serviço externo | Firebase, GCS, etc. |
| `KindBadGateway` | Gateway ruim | Resposta inválida de upstream |
| `KindUnavailable` | Serviço indisponível | Dependência fora do ar |
| `KindTimeout` | Timeout | Operação demorou demais |
| `KindServiceClosed` | Serviço fechado | Aplicação em shutdown |
| `KindInternal` | Erro interno | Fallback genérico |

---

## Criando Erros

### 1. Erro Simples (sem causa)

```go
var ErrPatientNotFound = apperr.New(
	apperr.KindNotFound,
	"patient_not_found",
	"patient not found",
)
```

### 2. Erro Encapsulando Outro (wrap)

```go
if err := repo.Save(ctx, patient); err != nil {
	return apperr.Wrap(
		apperr.KindExternal,
		"patient_save_failed",
		err,
	)
}
```

### 3. Erro Dinâmico

```go
func ValidateCPF(cpf string) error {
	if !isValidCPF(cpf) {
		return apperr.New(
			apperr.KindInvalidInput,
			"invalid_cpf",
			fmt.Sprintf("CPF %s is invalid", cpf),
		)
	}
	return nil
}
```

---

## Verificando Erros

### Verificar Kind

```go
if apperr.IsKind(err, apperr.KindNotFound) {
	// handle not found
}
```

### Extrair `apperr.Error`

```go
if ae, ok := apperr.As(err); ok {
	log.Info("app error", "kind", ae.Kind, "code", ae.Code)
}
```

---

## Tratamento na Camada HTTP

### Fluxo Padrão nos Handlers

```go
func (h *PatientHandler) GetByID(c *gin.Context) {
	// 1. Obter usuário autenticado
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok || currentUser == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// 2. Parsing de parâmetros (validação de fronteira)
	id := c.Param("id")
	if id == "" {
		common.RespondError(c, http.StatusBadRequest, "missing_patient_id", nil)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
		return
	}

	// 3. Chamada ao Service (retorna apperr.Error se falhar)
	patient, err := h.svc.GetByID(c.Request.Context(), currentUser, parsedID)
	if err != nil {
		// Delega tradução Kind → HTTP
		common.RespondAppError(c, err)
		return
	}

	// 4. Resposta de sucesso
	c.JSON(http.StatusOK, patient)
}
```

### Funções Auxiliares

#### `common.RespondError` — Erro Explícito

Para erros de **fronteira HTTP** (parsing, autenticação inicial):

```go
// Status e code são definidos explicitamente
common.RespondError(c, http.StatusBadRequest, "invalid_patient_id", err)
```

#### `common.RespondAppError` — Erro da Aplicação

Para erros vindos dos **Services/UseCases**:

```go
// Traduz apperr.Kind → Status HTTP automaticamente
common.RespondAppError(c, err)
```

### Mapeamento Kind → Status HTTP

Implementado em [`internal/http/api/handlers/common/error_mapper.go`](../../internal/http/api/handlers/common/error_mapper.go):

```go
func RespondAppError(c *gin.Context, err error) {
	ae, ok := apperr.As(err)
	if !ok {
		RespondError(c, http.StatusInternalServerError, "server_error", err)
		return
	}

	switch ae.Kind {
	case apperr.KindInvalidInput:
		RespondError(c, http.StatusBadRequest, ae.Code, err)
	case apperr.KindNotFound:
		RespondError(c, http.StatusNotFound, ae.Code, nil)
	case apperr.KindConflict:
		RespondError(c, http.StatusConflict, ae.Code, nil)
	case apperr.KindUnauthorized:
		RespondError(c, http.StatusUnauthorized, ae.Code, nil)
	case apperr.KindForbidden:
		RespondError(c, http.StatusForbidden, ae.Code, nil)
	case apperr.KindUnavailable, apperr.KindServiceClosed:
		RespondError(c, http.StatusServiceUnavailable, ae.Code, nil)
	case apperr.KindTimeout:
		RespondError(c, http.StatusGatewayTimeout, ae.Code, err)
	case apperr.KindBadGateway, apperr.KindExternal:
		RespondError(c, http.StatusBadGateway, ae.Code, err)
	default:
		RespondError(c, http.StatusInternalServerError, "server_error", err)
	}
}
```

---

## Logging Estruturado

Implementado em [`internal/http/api/handlers/common/respond_erros.go`](../../internal/http/api/handlers/common/respond_erros.go):

### Níveis de Log por Status

```go
switch {
case status >= 500:
	log.Error("handler_error", "status", status, "error", code, "err", err)
case status == 401 || status == 403:
	log.Info("handler_error", "status", status, "error", code)
case status == 400 && code == "invalid_input":
	log.Info("handler_error", "status", status, "error", code, "err", err)
default:
	log.Warn("handler_error", "status", status, "error", code, "err", err)
}
```

### Proteção de Dados Sensíveis

- Erros 5xx **não vazam detalhes** na resposta HTTP
- Detalhes internos ficam apenas nos logs
- Cliente recebe apenas `{"error": "server_error"}`

```go
if status >= 500 {
	c.JSON(status, gin.H{"error": code})
	return
}

// Para 4xx, inclui detalhes se disponíveis
if err != nil {
	c.JSON(status, gin.H{
		"error":   code,
		"details": err.Error(),
	})
	return
}
```

---

## Exemplo Completo

### Domínio/Service

```go
// internal/domain/entities/patient/errors.go
var ErrPatientNotFound = apperr.New(
	apperr.KindNotFound,
	"patient_not_found",
	"patient not found",
)

var ErrCPFConflict = apperr.New(
	apperr.KindConflict,
	"cpf_already_exists",
	"a patient with this CPF already exists",
)

// internal/app/services/patient/service_impl.go
func (s *service) Create(ctx context.Context, user *user.User, input CreateInput) (*patient.Patient, error) {
	// Policy check
	if err := s.policy.CanCreate(ctx, user); err != nil {
		return nil, err // já é apperr.Error (KindForbidden)
	}

	// Business validation
	existing, _ := s.repo.FindByCPF(ctx, input.CPF)
	if existing != nil {
		return nil, patient.ErrCPFConflict
	}

	// Create entity
	p, err := patient.New(input.CPF, input.FullName, ...)
	if err != nil {
		return nil, err // domain error (KindInvalidInput)
	}

	// Persist
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, apperr.Wrap(apperr.KindExternal, "patient_save_failed", err)
	}

	return p, nil
}
```

### Handler

```go
// internal/http/api/handlers/patient/patient_handler.go
func (h *PatientHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	log := applog.FromContext(ctx)
	log.Info("patient_create")

	// 1. Autenticação
	user, ok := middleware.GetCurrentUser(c)
	if !ok || user == nil {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// 2. Bind JSON
	var req CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_input", err)
		return
	}

	// 3. Parsing de fronteira (validações simples de formato)
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_birth_date", err)
		return
	}

	gender, err := shared.ParseGender(req.Gender)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid_gender", err)
		return
	}

	// 4. Montagem do input
	input := patientsvc.CreateInput{
		CPF:       req.CPF,
		FullName:  req.FullName,
		BirthDate: birthDate,
		Gender:    gender,
		// ...
	}

	// 5. Execução do service
	p, err := h.svc.Create(ctx, user, input)
	if err != nil {
		// Delega tradução Kind → HTTP
		common.RespondAppError(c, err)
		return
	}

	// 6. Resposta de sucesso
	c.Header("Location", "/patients/"+p.ID.String())
	c.JSON(http.StatusCreated, p)
}
```

### Resposta HTTP

**Sucesso:**
```http
HTTP/1.1 201 Created
Location: /patients/123e4567-e89b-12d3-a456-426614174000
Content-Type: application/json

{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "cpf": "12345678900",
  "full_name": "João Silva",
  ...
}
```

**Erro — CPF duplicado:**
```http
HTTP/1.1 409 Conflict
Content-Type: application/json

{
  "error": "cpf_already_exists"
}
```

**Erro — Validação:**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "invalid_cpf",
  "details": "CPF 123 is invalid"
}
```

**Erro — Não encontrado:**
```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": "patient_not_found"
}
```

**Erro — Sem permissão:**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": "forbidden"
}
```

---

## Checklist para Novos Erros

- [ ] Definir `Kind` apropriado (usar existente ou propor novo)
- [ ] Criar erro com `apperr.New()` ou `apperr.Wrap()`
- [ ] Definir `Code` estável (ex: `patient_not_found`)
- [ ] Adicionar mensagem descritiva (opcional)
- [ ] Documentar erro no domínio (comentário ou arquivo `errors.go`)
- [ ] Usar `common.RespondAppError()` no handler
- [ ] Verificar mapeamento `Kind → Status HTTP` em `error_mapper.go`

---

## Boas Práticas

### ✅ Fazer

- Usar `apperr.Error` em toda camada de aplicação/domínio
- Escolher `Kind` semântico, não status HTTP
- Manter `Code` estável (versionamento de API)
- Logar erros internos (5xx) sem expor detalhes ao cliente
- Usar `common.RespondAppError()` para erros de Service/UseCase
- Usar `common.RespondError()` para erros de fronteira HTTP

### ❌ Evitar

- Retornar status HTTP diretamente do domínio/service
- Criar novo `Kind` sem necessidade (reusar existentes)
- Expor stack traces ou mensagens internas em 5xx
- Fazer switch de erros de domínio nos handlers
- Misturar lógica de tradução HTTP fora de `error_mapper.go`

---

## Troubleshooting

### Erro não está sendo traduzido corretamente

1. Verificar se o erro é criado com `apperr.New()` ou `apperr.Wrap()`
2. Confirmar que o `Kind` está definido corretamente
3. Checar se `error_mapper.go` tem case para o `Kind`
4. Validar que handler usa `common.RespondAppError()`

### Log não está aparecendo

1. Verificar `LOG_LEVEL` (deve ser `debug`, `info`, `warn` ou `error`)
2. Confirmar que `applog.FromContext(ctx)` está sendo usado
3. Checar se middleware de logging está registrado

### Detalhes vazando em 5xx

1. Confirmar que `RespondError()` não inclui `err.Error()` em 5xx
2. Validar que erros 5xx retornam apenas `{"error": "server_error"}`
3. Checar logs estruturados para debugging interno

---

## Referências

- **ADR-006:** [apperr-kind-based-error-handling](adr/ADR-006-apperr-kind-based-error-handling.md)
- **Código:**
  - [`internal/app/apperr/`](../../internal/app/apperr/)
  - [`internal/http/api/handlers/common/error_mapper.go`](../../internal/http/api/handlers/common/error_mapper.go)
  - [`internal/http/api/handlers/common/respond_erros.go`](../../internal/http/api/handlers/common/respond_erros.go)
- **Exemplo:** [`internal/http/api/handlers/patient/patient_handler.go`](../../internal/http/api/handlers/patient/patient_handler.go)
