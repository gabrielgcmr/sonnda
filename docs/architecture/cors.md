<!-- docs/architecture/cors.md -->
# CORS (Cross-Origin Resource Sharing) - Implementação

## Visão Geral

Este documento descreve como CORS foi implementado no backend da Sonnda seguindo as melhores práticas profissionais para ambientes de desenvolvimento e produção.

## Cenários

### 1. **Desenvolvimento (dev)**
- Frontend rodando em `http://localhost:5173` (Vite)
- Backend rodando em `http://localhost:8080`
- CORS completamente ativado para facilitar desenvolvimento local
- Proxy do Vite **opcional** — backend já responde com headers CORS

**Default:**
```
CORS_ORIGINS=http://localhost:5173,http://localhost:3000
```

### 2. **Staging (staging)**
- Frontend em `https://app-staging.sonnda.com.br`
- API em seu próprio domínio (diferente)
- CORS ativado entre domínios

**Default:**
```
CORS_ORIGINS=https://app-staging.sonnda.com.br
```

### 3. **Produção (prod)**
- Frontend em `https://app.sonnda.com.br`
- API em seu próprio domínio (diferente)
- CORS ativado apenas para o domínio específico do frontend

**Default:**
```
CORS_ORIGINS=https://app.sonnda.com.br
```

**⚠️ MELHOR PRÁTICA PARA PRODUÇÃO:**
A topologia ideal seria:
- Frontend e API atrás do **mesmo host** (ex: `app.sonnda.com.br/api/`)
- Usando **reverse proxy** (Nginx, Traefik, Cloudflare, K8s ingress)
- Isso elimina CORS no browser e simplifica segurança

## Implementação

### Estrutura de Configuração

```
internal/config/cors.go
├── CORSConfig { AllowOrigins, AllowMethods, AllowHeaders, ExposeHeaders, AllowCredentials, MaxAge }
├── loadCORSConfig(appEnv string) → carrega automático por ambiente
└── parseCommaSeparatedList(s string) → parse de CORS_ORIGINS
```

### Middleware

```
internal/api/middleware/cors.go
├── SetupCors(cfg CORSConfig) → retorna gin.HandlerFunc
└── Integrado com github.com/gin-contrib/cors
```

### Integração no App

O middleware de CORS é registrado **primeiro** em `app.go`:

```go
r.Use(middleware.SetupCors(opts.CORSConfig))  // ← PRIMEIRO
r.Use(
    middleware.RequestID(),
    middleware.AccessLog(logger),
    middleware.Recovery(logger),
)
```

**Importante:** Estar **primeiro** garante que headers CORS sejam incluídos em **todas** as respostas, inclusive em erros (404, 401, 422, 500).

## Configuração por Variáveis de Ambiente

### Automática (recomendado)

Deixe os defaults automáticos conforme seu `APP_ENV`:

```bash
APP_ENV=dev    # → CORS_ORIGINS = http://localhost:5173,http://localhost:3000
APP_ENV=staging # → CORS_ORIGINS = https://app-staging.sonnda.com.br
APP_ENV=prod   # → CORS_ORIGINS = https://app.sonnda.com.br
```

### Manual (override)

Para sobrescrever, use variáveis de ambiente:

```bash
# Múltiplas origens (separadas por vírgula)
CORS_ORIGINS=https://app.sonnda.com.br,https://partner.example.com

# Permitir credentials (cookies, Authorization header)
CORS_CREDENTIALS=true

# TTL do preflight (em horas, inserção interna)
# (nota: o código usa maxAgeHours, conversão interna)
```

## Headers CORS Configurados

### Permitidos (Request)
- `Origin`
- `Content-Type`
- `Authorization` ← importante para JWT
- `Accept`
- `X-Request-ID`

### Expostos (Response)
- `Content-Length`
- `X-Request-ID`

### Metadados
- `Access-Control-Allow-Credentials: true` (padrão)
- `Access-Control-Max-Age: 12|24 horas` (conforme ambiente)
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS`

## Fluxo de Requisições CORS

### 1. **Preflight (OPTIONS)**

O browser envia automaticamente:
```http
OPTIONS /v1/me HTTP/1.1
Origin: http://localhost:5173
Access-Control-Request-Method: GET
Access-Control-Request-Headers: Authorization
```

Resposta do backend:
```http
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Authorization, Accept, X-Request-ID
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 43200
```

### 2. **Requisição Real (GET/POST/etc)**

```http
GET /v1/me HTTP/1.1
Origin: http://localhost:5173
Authorization: Bearer eyJ...
```

Resposta do backend (sucesso):
```http
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Credentials: true
Content-Type: application/json

{ "id": "user123", ... }
```

Resposta do backend (erro 404):
```http
HTTP/1.1 404 Not Found
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Credentials: true
Content-Type: application/problem+json

{ "type": "urn:sonnda:error:resource-not-found", ... }
```

**⚠️ CRÍTICO:** Headers CORS devem estar presentes **inclusive em respostas de erro**.

## Diferença: Proxy Vite vs CORS Backend

### Proxy Vite (frontend)
```javascript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, '/v1')
    }
  }
}
```

**Quando usar:**
- ✅ Desenvolvimento local sem CORS overhead
- ✅ Simplifica setup de frontend
- ✅ O browser vê requisições como same-origin

**Limitações:**
- ❌ Não funciona em produção (apenas dev local)
- ❌ Máscara o cenário real de domínios diferentes

### CORS Backend (recomendado)
**Quando usar:**
- ✅ Funciona em todos os ambientes (dev, staging, prod)
- ✅ Simula o cenário real de produção
- ✅ Não mascara problemas de CORS localmente

**Conclusão:**
Em desenvolvimento, **ambos podem coexistir**:
- Backend com CORS configurado
- Frontend **opcionalmente** com proxy Vite

Isso oferece flexibilidade: desenvolvedores podem desativar o proxy no Vite e usar o CORS real do backend.

## Tratamento de Erros com CORS

### Antes (❌ problema)
```
GET /v1/me → 404 Not Found
Sem Access-Control-Allow-Origin

Browser bloqueia a resposta:
"Access to XMLHttpRequest has been blocked by CORS policy"
```

### Depois (✅ correto)
```
GET /v1/me → 404 Not Found
Access-Control-Allow-Origin: http://localhost:5173
Content-Type: application/problem+json

Frontend consegue ler 404 e redirecionar para /onboarding
```

## Checklista de Configuração

- [ ] Verificar `APP_ENV` está correto (dev|staging|prod)
- [ ] Testar preflight OPTIONS com devtools
- [ ] Verificar se `Authorization` header funciona
- [ ] Testar erro 404 — deve ter headers CORS
- [ ] Testar erro 401 — deve ter headers CORS
- [ ] Em produção, considerar reverse proxy ao invés de CORS

## Referências

- [RFC 9110 - HTTP Semantics (CORS)](https://httpwg.org/specs/rfc9111.html)
- [MDN - CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [gin-contrib/cors - Documentação](https://github.com/gin-contrib/cors)
