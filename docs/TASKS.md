# Tasks — GeevheitNeural

> Fonte: [`/sprit01`](../sprit01). Legenda: `[ ]` a fazer · `[/]` em andamento · `[x]` feito.
> Regra: **parar ao fim de cada fase e aguardar aprovação do dono.**

## Fase 0 — Plano e decisões *(atual)*
- [x] T0-01 Ler o spec e montar PRD, Architecture, System Design, Memory, Rules e Tasks.
- [x] T0-02 **Dono aprova o plano** e responde às [perguntas em aberto](./MEMORY.md#perguntas-em-aberto).

## Fase 1 — Fundação *(concluída - aguardando aprovação do dono para Fase 2)*
Definição de pronto: `make check` verde no local e no CI; README explica como rodar do zero.

| ID | Tarefa | Aceite |
|---|---|---|
| [x] F1-01 | Esqueleto do monorepo: `api/ web/ db/ docs/`, `Makefile`, `.editorconfig`, `.gitignore`, `git init`. | `make help` lista os alvos. |
| [x] F1-02 | Ambiente local com Supabase CLI: cadastro desligado, MFA TOTP ligado, `.env.example` sem segredos. | `make db-up` sobe Postgres + Auth. |
| [x] F1-03 | Migration `00001_schema`: schema `geev`, enums, todas as tabelas do modelo, dinheiro em centavos, `updated_at`, `system_state` singleton. | `goose up/down` limpos. |
| [x] F1-04 | Migration `00002_security`: RLS habilitada e forçada em **todas** as tabelas, `geev.is_owner()`, `geev.can_execute()`, roles `geev_api`/`geev_intake`, triggers de imutabilidade. | pgTAP: sem JWT ⇒ 0 linhas; aal1 ⇒ 0 linhas; UPDATE em `approvals` ⇒ erro. |
| [x] F1-05 | Domínio: `money`, `card` (estágios, transições, portões), `approval`, `system` (ExecutionGuard). | Testes de unidade: transição inválida falha; portão só sai via `Decide`; guard nega se pausado ou inativo. |
| [x] F1-06 | Casos de uso: `DecideGate`, `PauseSystem`, `ResumeSystem`, `GetMoneyWaiting` + portas. | Testes com fakes: aprovação gera `approval` + `card_event` + `audit` na mesma UoW; agente não consegue chamar `DecideGate`. |
| [x] F1-07 | Adapter Postgres: sqlc + pgx, `UnitOfWork` que aplica role/claims por transação, repositórios. | Testes de integração contra o banco local. |
| [x] F1-08 | Adapter auth: verificador JWT (JWKS com cache), exige `aal2` e `sub = OWNER_USER_ID`. | Testes: token expirado, aal1, outro `sub` ⇒ 401/403. |
| [x] F1-09 | API base: config por env, `slog` JSON sem PII, request id, recover, CORS restrito, timeouts, graceful shutdown, `/healthz`, `/readyz`, rotas de F1-06. | `curl` documentado no README. |
| [x] F1-10 | `api/openapi.yaml` v1 com os endpoints da fase 1. | Validação do spec no CI. |
| [x] F1-11 | `api/Dockerfile` multi-stage, imagem distroless, usuário não-root. | `docker build` ok, imagem < 30 MB (23.7 MB). |
| [x] F1-12 | Web base: Next.js static export, Tailwind, TanStack Query, tema claro/escuro, layout 360 px, login + TOTP, cliente da API. | Login com 2FA funcionando contra o Supabase local. |
| [x] F1-13 | CI (GitHub Actions): lint + testes de api, db (pgTAP) e web. | Pipeline verde. |
| [x] F1-14 | README de execução + resumo da fase em `docs/CHANGELOG.md`. | Rodar do zero seguindo só o README. |

## Fase 2 — Painel *(concluída - aguardando aprovação do dono para Fase 3)*
- [x] F2-01 Shell da PWA: navegação inferior, contador **Dinheiro parado** fixo, faixa "Sistema pausado".
- [x] F2-02 Hoje: 3 decisões, recomendação de 1 linha e ação direta.
- [x] F2-03 Quadro Kanban: colunas por estágio, drawer de detalhe, histórico, aprovar / pedir info / recusar.
- [x] F2-04 Meus apps: CRUD, custo × receita, alerta "custa mais do que rende".
- [x] F2-05 Finanças: lançamentos manuais, importação CSV, visões por app/cliente/serviço, sugestão de consolidação.
- [x] F2-06 Conexões: lista fixa, tudo desligado, `secret_ref` opcional, escopo leitura.
- [x] F2-07 Botão Parar/Retomar com confirmação + `audit_log`.
- [x] F2-08 SSE `/v1/stream` + atualização do contador em tempo real.
- [x] F2-09 PWA instalável: manifest, ícones, service worker (offline para leitura).
- [x] F2-10 Teste de fluxo: cadastrar apps, lançar custos e ver lucro em < 5 min.

## Fase 3 — Comando e risco *(concluída)*
- [x] F3-01 Parser de comandos (aprovar, recusar, status, quanto gastei/lucro) + testes de ambiguidade.
- [x] F3-02 Confirmação de ações destrutivas (token de uso único).
- [x] F3-03 Barra de comando global + ditado Web Speech API pt-BR.
- [x] F3-04 Ficha de risco: modelo, preenchimento manual, exibição obrigatória antes de aprovar entrega.

## Fase 4 — Fluxo comercial *(concluída)*
- [x] F4-01 `POST /v1/intake` com validação, honeypot, rate limit e role `geev_intake`.
- [x] F4-02 Propostas: criar a partir do cartão, enviar link de aceite (token com hash).
- [x] F4-03 Página pública de aceite → move para *Prompt para aprovar*.
- [x] F4-04 Export e expurgo de dados por cliente (LGPD).

## Fase 5 — Contratos dos agentes *(concluída)*
- [x] F5-01 OpenAPI `/v1/agent/*`.
- [x] F5-02 Endpoints implementados atrás do ExecutionGuard (respondem `423` enquanto `ativo=false`).
- [x] F5-03 Tokens de agente (hash), limite de gasto, escopo mínimo.
- [x] F5-04 Testes: agente inativo ou sistema pausado ⇒ nenhuma escrita; agente nunca aprova.
