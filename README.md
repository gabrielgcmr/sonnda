# Sonnda API

API backend da plataforma **Sonnda**, focada em atenção primária à saúde, com suporte a:

- cadastro e gerenciamento de pacientes;
- upload e processamento de exames laboratoriais;
- extração automática de dados estruturados via **Google Cloud Document AI**;
- armazenamento seguro em **PostgreSQL (Supabase)**;
- arquitetura limpa (hexagonal), preparada para evolução e escala.

> Atenção: este repositório **não** deve conter dados reais de pacientes nem arquivos de configuração sensíveis (`.env`). Veja a seção [Boas práticas de segurança](#boas-práticas-de-segurança).

---

## Sumário

- [Arquitetura](#arquitetura)
- [Stack Tecnológico](#stack-tecnológico)
- [Estrutura de Pastas](#estrutura-de-pastas)
- [Pré-requisitos](#pré-requisitos)
- [Configuração](#configuração)
  - [Variáveis de Ambiente](#variáveis-de-ambiente)
  - [Configuração do Supabase / PostgreSQL](#configuração-do-supabase--postgresql)
  - [Configuração do Document AI](#configuração-do-document-ai)
- [Como Rodar o Projeto](#como-rodar-o-projeto)
- [Fluxos Principais](#fluxos-principais)
  - [Autenticação](#autenticação)
  - [Upload e Processamento de Exames](#upload-e-processamento-de-exames)
- [Boas Práticas de Segurança](#boas-práticas-de-segurança)
- [Roadmap](#roadmap)
- [Contribuição](#contribuição)
- [Licença](#licença)

---

## Arquitetura

O projeto segue princípios de **Clean Architecture / Arquitetura Hexagonal**, separando claramente:

- **Domínio (`core/domain`)**  
  Entidades e regras de negócio (ex.: `Patient`, `LabReport`, etc.).

- **Ports (`core/ports`)**  
  Interfaces que descrevem o que o domínio espera de:
  - Repositórios (persistência);
  - Serviços externos (ex.: Document AI);
  - Casos de uso.

- **Adapters (`internal/adapters`)**  
  Implementações concretas:
  - HTTP (handlers, middlewares, DTOs);
  - Supabase/PostgreSQL (repositórios);
  - Document AI (cliente de processamento de documentos).

- **Config / Infra (`internal/config`, `internal/infra`)**  
  Leitura de variáveis de ambiente, logger, conexão com banco, etc.

---

## Stack Tecnológico

- **Linguagem:** Go (Golang)
- **Banco de dados:** PostgreSQL (gerenciado via **Supabase**)
- **ORM / Driver:** `pgx` / `pgxpool`
- **Processamento de documentos:** Google Cloud **Document AI**
- **Autenticação:** JWT (tokens de acesso)
- **Containerização:** Docker / docker-compose (para ambiente local)
- **Arquitetura:** Hexagonal / Clean Architecture

---

## Estrutura de Pastas

Resumo da estrutura (pode ajustar conforme seu projeto atual):

```text
.
├── cmd/
│   └── api/
│       └── main.go                    # Ponto de entrada da API
├── internal/
│   ├── config/                        # Leitura de env, configuração geral
│   ├── core/
│   │   ├── domain/                    # Entidades de domínio (Patient, LabReport, etc.)
│   │   ├── ports/                     # Interfaces (repositories, services)
│   │   └── usecases/                  # Casos de uso da aplicação (orquestram o domínio)
│   └── adapters/
│   │   ├── inbound/                   # Tudo que vem de fora da aplicação
│   │   │   ├── cli/                   # Entradas CLI
│   │   │   ├── http/                  # Entradas HTTP, roteamento e middlewares
│   │   ├── outbound/                  # Tudo que vai pra fora da aplicação
│   │   │   ├── auth/                  # Implementação e ajustes do sistema da autenticação
│   │   │   ├── authorization/         # Regras de autorização (RBAC, checagem de permissões)
│   │   │   ├── database/              # Implementação de repositórios usando Supabase/Postgres
│   │   │   ├── storage/               # Implementação do storage
│   │   │   ├── external/              # Chamadas a APIs externas.
│   │   │   │   └── documentai/        # Cliente para Google Document AI
├── samples/                           # Exemplos de payloads / exames (sempre dados fictícios/anonimizados)
├── .env.example                       # Exemplo de variáveis de ambiente (sem segredos)
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
