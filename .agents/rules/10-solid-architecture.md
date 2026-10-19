---
trigger: always_on
---

# GeevheitNeural — SOLID e arquitetura

## Backend Go (`/api`)
- Camadas: `domain` → `app` → `adapters` → `cmd` (composition root). Dependências só apontam para dentro.
- `internal/domain`: puro. Proibido importar `database/sql`, `net/http`, `pgx`, `os`, `time.Now()` direto (usar `Clock`).
- `internal/app`: um caso de uso por arquivo; declara as **portas** (interfaces) que usa. Nunca importa `adapters`.
- `internal/adapters`: implementam portas. Handlers HTTP só traduzem request ⇄ caso de uso ⇄ response.
- Injeção de dependência manual em `cmd/geevheit-api/main.go`. Sem frameworks de DI.

### SOLID na prática
- **S:** uma razão para mudar por tipo/arquivo. Handler não tem regra de negócio; repositório não tem regra de negócio.
- **O:** extensões por novas implementações (comando, conector, fonte de risco), sem editar `switch` central.
- **L:** toda implementação de porta passa no mesmo teste de contrato (`*_contract_test.go`).
- **I:** interfaces pequenas, definidas no consumidor (`CardReader`, `CardWriter`, `ProgressRecorder`…).
- **D:** `app` depende de abstrações; detalhes (Postgres, Supabase, SSE) são plugados no `main`.
- Privilégio pelo tipo: `DecideGate` recebe `domain.OwnerPrincipal`; agentes usam `domain.AgentPrincipal` e nunca chegam a esse caminho.

### Go
- Go ≥ 1.22, roteamento com `net/http` (`mux.HandleFunc("POST /v1/...")`).
- Erros: `errors.Is/As`, erros de domínio sentinela (`card.ErrGateRequiresOwner`, `system.ErrPaused`). Nunca `panic` fora de `main`.
- `context.Context` como primeiro parâmetro em tudo que faz I/O.
- Logs com `log/slog` JSON, sem dados pessoais.
- SQL só em `adapters/postgres/queries/*.sql` (sqlc). Nada de SQL concatenado.
- Testes: tabela (`t.Run`), `-race`, fakes das portas em `app`, integração em `adapters/postgres`.

## Banco (`/db`)
- Migrations goose em SQL puro com `Up` e `Down`. Nunca editar migration já aplicada; criar nova.
- Toda tabela nova: `ENABLE` + `FORCE ROW LEVEL SECURITY` e policies **na mesma migration**, com teste pgTAP.
- Tabelas imutáveis: só `SELECT`/`INSERT` + trigger que bloqueia `UPDATE`/`DELETE`.
- Segredos nunca no banco — só `secret_ref` (nome da variável/cofre).

## Frontend (`/web`)
- Next.js App Router com `output: 'export'`; nada de Server Actions/SSR.
- Organização por feature: `src/features/<feature>/{components,hooks,api}`; compartilhado em `src/shared`.
- Componentes de apresentação não buscam dados; hooks TanStack Query encapsulam a API.
- Toda chamada passa por `src/shared/api-client` (um ponto para auth, erros e base URL).
- Mobile-first: validar em 360 px; tema claro e escuro via CSS variables.
- Nenhum segredo no bundle. Só `NEXT_PUBLIC_SUPABASE_URL`, a chave pública do Supabase e `NEXT_PUBLIC_API_URL`.
- IDs únicos e descritivos em elementos interativos (para testes).
