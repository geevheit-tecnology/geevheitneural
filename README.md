# GeevheitNeural

Central de operação de software house, mobile-first. Um único dono decide, aprova e dá ordens pelo celular; agentes de IA ficam preparados, mas desligados.

---

## 🏗️ Arquitetura e Monorepo

O repositório é organizado em quatro pastas principais:
- `/api`: Backend em Go (≥ 1.22), Clean Architecture / Hexagonal / SOLID, stdlib `net/http`, `pgx/v5`, `sqlc`, Distroless Docker image (< 30 MB).
- `/web`: Frontend PWA em Next.js (App Router, static export `output: 'export'`), TypeScript, Tailwind CSS v4, TanStack Query, mobile-first (360px) com temas claro e escuro.
- `/db`: Banco de dados PostgreSQL 15+ gerenciado via Supabase CLI local com migrations puras SQL via Goose e testes de segurança pgTAP.
- `/docs`: Documentação viva do sistema (PRD, Arquitetura com ADRs, System Design, Tasks, Memory e Changelog).

---

## ⚡ Pré-requisitos

- **Go** ≥ 1.22
- **Node.js** ≥ 20 e **npm** ≥ 10
- **Docker Desktop** ativo
- **Supabase CLI** (`npm install -g supabase` ou via brew)
- **Goose** CLI para migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- **sqlc** CLI para compilação de SQL em Go (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

---

## 🚀 Como Executar do Zero

### 1. Iniciar o Banco de Dados Local (Supabase)
O projeto utiliza instâncias isoladas com portas configuradas para não conflitar com outros serviços locais (DB: `54332`, API Gateway: `54331`):

```bash
make db-up
```

### 2. Aplicar as Migrations do Banco
Executa o schema `geev`, tabelas, restrições, triggers e policies de Row-Level Security (RLS):

```bash
make migrate-up
```

Para verificar o status das migrations:
```bash
make migrate-status
```

### 3. Rodar os Testes Automatizados

**Testes de Segurança RLS (pgTAP):**
```bash
cd db && npx supabase test db
```

**Testes do Backend Go (com detector de race conditions):**
```bash
make api-test
```

### 4. Iniciar a API Go
Crie o arquivo de ambiente a partir do exemplo:
```bash
cp .env.example api/.env
```

Execute a API:
```bash
make api-run
```
A API estará acessível em `http://localhost:8080`.
Endpoints de saúde:
- `http://localhost:8080/healthz` (liveness)
- `http://localhost:8080/readyz` (readiness com ping no PostgreSQL)

### 5. Iniciar o Frontend Web (PWA)
Em outro terminal:
```bash
cd web
npm install
npm run dev
```
Acesse `http://localhost:3000` no navegador (ou abra no modo emulação móvel com largura de 360px).

---

## 🐳 Build de Contêiner da API

Para gerar a imagem Docker enxuta e segura (não-root, multi-stage, < 25 MB):
```bash
docker build -t geevheit-api:latest api/
```

---

## 🔒 Princípios e Invariantes do Sistema

1. **API Go como ponto único de entrada**: Nenhuma ação ou mutação de estado ignora a API.
2. **Aprovação exclusiva do dono**: Portões de decisão (*Proposta*, *Prompt*, *Entrega*) só avançam por decisão explícita do dono via token AAL2 TOTP. Agentes nunca aprovam portões.
3. **ExecutionGuard**: Se o sistema estiver pausado (`system.pausado=true`) ou conector inativo (`ativo=false`), nenhuma execução externa ocorre.
4. **Imutabilidade**: As tabelas `approvals`, `card_events` e `audit_log` são estritamente append-only e protegidas contra `UPDATE` e `DELETE` por triggers.
5. **Finanças Seguras**: Valores monetários são sempre trafegados e persistidos em centavos de Real (`int64`/`bigint`), nunca em ponto flutuante.
