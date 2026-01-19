# Autenticação

## Visão Geral

A autenticação do Sonnda utiliza **Firebase Authentication**. O fluxo é:

1. **Cliente** (mobile/web) autentica diretamente com Firebase
2. Firebase retorna um **ID Token** (JWT)
3. Cliente envia o token no header `Authorization: Bearer <token>` para a API
4. **API** valida o token via middleware e extrai a identidade do usuário

```
┌─────────────┐     ┌──────────────┐     ┌───────────────────┐
│   Cliente   │────▶│   Firebase   │────▶│ Retorna ID Token  │
│   (Mobile)  │     │  Auth API    │     │                   │
└─────────────┘     └──────────────┘     └───────────────────┘
       │                                           │
       │         Authorization: Bearer <token>     │
       ▼                                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      sonnda-api                             │
│         AuthMiddleware → FirebaseAuthService                │
│              (Valida token e extrai identity)               │
└─────────────────────────────────────────────────────────────┘
```

---

## Arquivos-chave (API)

| Camada     | Arquivo                                                    |
|------------|------------------------------------------------------------|
| Middleware | `internal/adapters/inbound/http/middleware/auth.go`        |
| Service    | `internal/adapters/outbound/integrations/firebase_auth.go` |
| Port       | `internal/domain/ports/auth.go`                            |
| Model      | `internal/domain/model/identity/identity.go`               |

---

## Endpoints Firebase (Client-Side)

> ⚠️ Esses endpoints são chamados **diretamente pelo cliente**, não pela API.

Base URL: `https://identitytoolkit.googleapis.com/v1`

### Criar Conta (Sign Up)

```
POST /accounts:signUp?key=<FIREBASE_API_KEY>
```

**Request Body:**
```json
{
  "email": "usuario@exemplo.com",
  "password": "senhaSegura123",
  "returnSecureToken": true
}
```

**Response (200 OK):**
```json
{
  "idToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "email": "usuario@exemplo.com",
  "refreshToken": "AMf-vBw...",
  "expiresIn": "3600",
  "localId": "abc123xyz"
}
```

**Erros comuns:**

| Código Firebase         | Descrição                        |
|-------------------------|----------------------------------|
| EMAIL_EXISTS            | Email já cadastrado              |
| WEAK_PASSWORD           | Senha com menos de 6 caracteres  |
| INVALID_EMAIL           | Formato de email inválido        |

---

### Login (Sign In with Password)

```
POST /accounts:signInWithPassword?key=<FIREBASE_API_KEY>
```

**Request Body:**
```json
{
  "email": "usuario@exemplo.com",
  "password": "senhaSegura123",
  "returnSecureToken": true
}
```

**Response (200 OK):**
```json
{
  "idToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "email": "usuario@exemplo.com",
  "refreshToken": "AMf-vBw...",
  "expiresIn": "3600",
  "localId": "abc123xyz",
  "registered": true
}
```

**Erros comuns:**

| Código Firebase          | Descrição                    |
|--------------------------|------------------------------|
| EMAIL_NOT_FOUND          | Usuário não existe           |
| INVALID_PASSWORD         | Senha incorreta              |
| USER_DISABLED            | Conta desativada             |
| INVALID_LOGIN_CREDENTIALS| Email ou senha inválidos     |

---

### Refresh Token

```
POST https://securetoken.googleapis.com/v1/token?key=<FIREBASE_API_KEY>
```

**Request Body (form-urlencoded ou JSON):**
```json
{
  "grant_type": "refresh_token",
  "refresh_token": "AMf-vBw..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": "3600",
  "token_type": "Bearer",
  "refresh_token": "AMf-vBw...",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "abc123xyz",
  "project_id": "sonnda-project"
}
```

**Erros comuns:**

| Código Firebase   | Descrição                     |
|-------------------|-------------------------------|
| TOKEN_EXPIRED     | Refresh token expirado        |
| USER_DISABLED     | Conta desativada              |
| INVALID_REFRESH_TOKEN | Refresh token inválido    |

---

## Validação na API (Server-Side)

### Header de Autenticação

Todas as rotas protegidas exigem:

```
Authorization: Bearer <idToken>
```

O `idToken` é o JWT retornado pelo Firebase nos endpoints acima.

### Erros da API

| Código AppError       | HTTP Status | Descrição                          |
|-----------------------|-------------|------------------------------------|
| `AUTH_REQUIRED`       | 401         | Header Authorization ausente       |
| `AUTH_TOKEN_INVALID`  | 401         | Token inválido ou malformado       |
| `AUTH_TOKEN_EXPIRED`  | 401         | Token expirado                     |
| `ACCESS_DENIED`       | 403         | Usuário sem permissão para recurso |

**Exemplo de resposta de erro:**
```json
{
  "error": {
    "code": "AUTH_TOKEN_INVALID",
    "message": "Token de autenticação inválido"
  }
}
```

---

## Fluxo Completo de Autenticação

1. **Novo usuário:**
   - Cliente chama `signUp` → recebe `idToken`
   - Cliente chama `POST /api/v1/register` com `Authorization: Bearer <idToken>` para completar onboarding

2. **Usuário existente:**
   - Cliente chama `signInWithPassword` → recebe `idToken`
   - Cliente usa `idToken` em todas as requisições protegidas

3. **Token expirado:**
   - Cliente chama `token` com `refresh_token` → recebe novo `idToken`
   - Cliente continua usando a API normalmente
