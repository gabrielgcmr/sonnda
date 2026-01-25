# ADR-007 — Separação entre templates e arquivos estáticos (sem pasta /assets)

## Status

**Status:** Aceito  
**Data:** 2026-01  
**Contexto:** Sonnda WEB — estrutura de pasta de web

## Contexto

Este projeto utiliza renderização server-side (Go + templ) com HTMX.
Nesse modelo, existem dois tipos de artefatos fundamentalmente diferentes:

* **Templates**: arquivos que participam da lógica de renderização HTML no servidor.
* **Arquivos estáticos**: arquivos públicos servidos diretamente ao browser (CSS, JS, imagens, fontes).

É comum em projetos web genéricos agrupar tudo em uma pasta `assets/`. No entanto, esse padrão mistura responsabilidades distintas e tende a causar confusão em projetos SSR, especialmente quando há automação, cache agressivo ou uso de agentes/IA.

## Decisão

* **Não utilizar uma pasta `assets/` no repositório.**
* Separar claramente por responsabilidade:

```
/templates   → templates SSR (.templ), componentes de layout e features
/public      → arquivos estáticos (css, js, images, fonts)
```

* Arquivos em `/templates` **não devem** ser servidos diretamente ao browser.
* Arquivos em `/public` **devem** ser servidos diretamente.

### Rota pública padronizada

Os arquivos estáticos serão expostos via URL:

```
/static/*  → mapeado para /public/*
```

O nome da rota (`/static`) é uma decisão de infraestrutura e **não implica** na existência de uma pasta `assets/` no código.

## Consequências

### Positivas

* Clareza arquitetural: UI renderizada ≠ arquivos públicos.
* Menor risco de exposição acidental de templates.
* Facilita cache, CDN e versionamento de estáticos.
* Evita que agentes/PRs assumam automaticamente a existência de `/assets`.

### Negativas

* Exige documentação explícita para evitar correções automáticas incorretas.
* Pode contrariar expectativas de ferramentas ou assistentes genéricos.

## Diretrizes para contribuições

* **Não criar** diretório `assets/`.
* **Não alterar caminhos** para `/assets/...` em templates ou scripts.
* Sempre usar URLs no formato `/static/...` para arquivos públicos.
* Templates devem referenciar apenas `/static` como origem de CSS/JS/imagens.

## Observação

Caso uma ferramenta ou agente sugira ajustes para “bater com assets”, essa sugestão deve ser considerada **incorreta para este projeto** e ajustada para respeitar esta decisão arquitetural.
