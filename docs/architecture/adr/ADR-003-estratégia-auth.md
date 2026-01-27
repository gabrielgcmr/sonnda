ADR-00X — Estratégia de Autenticação: Auth0 com Bearer Token (API) e Sessão via Cookie (Web)

Status: Aceito
Data: 2026-01-26
Decisores: Equipe Sonnda
Contexto: Autenticação e Autorização (Web + API + Mobile)

Contexto

A plataforma Sonnda possui múltiplos canais de acesso com características arquiteturais distintas:

API HTTP (JSON) consumida por aplicações externas e mobile.

Web SSR construída com Go + templ + HTMX, sem uso extensivo de JavaScript no cliente.

Aplicação Mobile (React Native / Expo), que consome a API diretamente.

Inicialmente, a autenticação foi implementada utilizando Firebase Auth, porém surgiram dificuldades arquiteturais na integração com o frontend Web SSR, especialmente relacionadas a:

Uso de tokens armazenados no cliente (LocalStorage).

Necessidade de injeção manual de headers Authorization em requisições HTMX.

Complexidade adicional para manter cookies HttpOnly seguros sem gambiarras.

Conflito entre o modelo token-first (SPA/mobile) e session-first (SSR).

Esses problemas não representam uma falha de segurança do Firebase Auth, mas sim uma incompatibilidade entre o modelo de autenticação e a arquitetura SSR adotada.

Decisão

Foi decidido adotar Auth0 como Identity Provider (IdP), utilizando dois mecanismos de autenticação distintos, de acordo com o canal de acesso:

1. API (incluindo Mobile)

Autenticação via Bearer Token (JWT).

Tokens emitidos pelo Auth0.

Validação no backend Go via JWKS (OIDC padrão).

Middleware existente de autenticação por Bearer é mantido.

Modelo stateless, escalável e adequado para mobile.

2. Web SSR (templ + HTMX)

Autenticação via OAuth2 Authorization Code Flow com Auth0.

Backend troca o authorization_code por tokens.

Criação de sessão server-side persistida em Redis (Upstash).

Cookie HttpOnly, Secure e SameSite=Lax contendo apenas o session_id.

O HTMX envia cookies automaticamente, sem necessidade de JavaScript adicional.

Ambos os fluxos convergem para o mesmo conceito interno de Identity, injetado no contexto da requisição, permitindo reutilização de políticas de autorização e regras de negócio.

Consequências
Positivas

Arquitetura alinhada com padrões Web modernos (OIDC).

Excelente integração com SSR + HTMX, sem hacks no frontend.

Redução do risco de XSS, já que tokens não ficam acessíveis ao JavaScript no Web.

Reaproveitamento quase total da infraestrutura atual da API.

Separação clara de responsabilidades:

Web → Sessão (stateful)

API/Mobile → Token (stateless)

Melhor clareza conceitual para futuros desenvolvedores.

Negativas / Trade-offs

Dupla estratégia de autenticação (cookie + bearer) aumenta levemente a complexidade conceitual.

Auth0 pode gerar custos maiores caso o número de usuários gratuitos cresça significativamente.

Experiência de login no mobile ocorre via navegador externo (padrão OAuth), menos “nativa” que Firebase.

Alternativas Consideradas
Firebase Auth

Excelente para mobile e SPAs.

Integração com SSR exige criação manual de session cookies e fluxos adicionais.

Aumenta a complexidade de manutenção no Web SSR.

Supabase Auth

Boa integração com mobile e backend.

Menor maturidade e flexibilidade em cenários B2B/Enterprise.

Decisão adiada para possível reavaliaç
ão futura.

Unificar tudo em Bearer Token

Exigiria armazenar tokens no cliente Web ou injetar headers manualmente no HTMX.

Aumentaria o risco de XSS ou a complexidade do frontend.

Rejeitado por não respeitar a natureza do SSR.

Decisão Final

A adoção de Auth0 com Bearer Token para API/Mobile e Sessão via Cookie para Web SSR é considerada a solução mais alinhada com os objetivos de segurança, manutenibilidade e clareza arquitetural do projeto Sonnda no momento.

Essa decisão poderá ser reavaliada caso:

O custo do Auth0 se torne impeditivo.

A arquitetura Web mude para SPA.

Um Identity Provider alternativo ofereça melhor custo-benefício sem comprometer o SSR.