# Changelog — GeevheitNeural

Todas as mudanças notáveis deste projeto serão documentadas neste arquivo seguindo o formato [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/) e [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Sprint 03 - Fase 12] - Identidade Visual Proprietária & Branding — 2026-10-21

### Adicionado
- **Logotipo e Ícones Customizados GeevheitNeural:**
  - Substituição completa de todos os favicons e logos genéricos do Next.js / Vercel.
  - Implementação do ícone orbital "G" constelação futurista em alta definição (1024x1024).
  - Ícones PWA multi-resolução gerados: `icon-192.png`, `icon-512.png`, `apple-touch-icon.png`, `favicon-32x32.png` e container Windows ICO multi-formato (`favicon.ico`).
  - Atualização do `manifest.json` com apontamento para os novos ícones PNG maskable.
  - Integração visual no Header do Dashboard mobile-first ao lado do título da aplicação.
  - Integração visual no topo do `LoginForm` de autenticação segura do Dono.

---

## [Sprint 03 - Fase 11] - Motor de Cobrança PIX e Conciliação Financeira — 2026-10-21

### Adicionado
- **Motor de Cobrança PIX Oficial (BR Code / EMV-QRCPS-MPM):**
  - Geração determinística de payload copia-e-cola EMV com cálculo de CRC16 (polinômio 0x1021 CCITT-FALSE).
  - Tabela `geev.payment_orders` com Row Level Security (RLS) forçada e triggers automáticos de timestamp.
  - Endpoints `POST /v1/cards/{id}/pix` e `GET /v1/cards/{id}/pix` exclusivos do Dono.
  - Endpoint público `POST /v1/payments/pix/webhook` para confirmação automática de pagamentos.
  - Conciliação bancária automática: a confirmação do webhook liquida a ordem, registra receita em `financial_entries` na mesma UoW e dispara evento de histórico no cartão.
  - Interface visual integrada na `CardDetailSheet` com QR Code dinâmico, código Copia-e-Cola e botão de 1 toque para copiar.

---

## [Sprint 02] - Fábrica Autônoma de Software (Fases 6 a 10) — 2026-10-18

### Adicionado
- **Fase 6: Comando Único "Construir" & Pipeline de Agentes:**
  - Máquina de estados sequencial do pipeline com os 7 papéis: `Requisitos` → `Arquiteto` → `Dev` → `QA` → `Segurança` → `Publicador` → `Suporte`.
  - Abstração `AgentRunner` plugável, conectada a `MockAgentRunner` com computação determinística de custos e artefatos.
  - Endpoint `POST /v1/cards/{id}/pipeline/start` e `GET /v1/cards/{id}/pipeline`.
  - Proteção por teto de gastos configurável por projeto (`spend_limit_cents`), interrompendo o pipeline com status `paused_spend_limit` ao estourar o limite.
  - Invariante dos Portões: o agente Publicador avança para `entrega_aprovar` (Portão 3) e NUNCA faz deploy direto sem decisão explícita do Dono.
- **Fase 7: Especificação Executável & Validação de Escopo BDD:**
  - Agente de Requisitos gera User Stories e Critérios de Aceite no formato BDD (Given-When-Then).
  - Portal seguro do cliente em `/escopo?token=...` com aprovação pública imutável registrada em `card_events`.
  - Garantia de sinal de entrada financeiro antes de liberar o início do desenvolvimento.
- **Fase 8: Biblioteca de Playbooks & Módulos Catalogados:**
  - Catálogo nativo com 3 playbooks de produtos (`SaaS B2B Multi-tenant`, `App Mobile Flutter / PWA`, `Landing Page de Alta Conversão`).
  - 4 módulos reutilizáveis catalogados e revisados (`auth_supabase_totp`, `payment_webhook_idempotent`, `audit_trail_immutable`, `kanban_bottom_sheet_mobile`).
  - Endpoints `GET /v1/playbooks` e `GET /v1/modules` com seed determinístico na inicialização.
- **Fase 9: Qualidade Automática & Entrega Segura:**
  - Agente QA e Segurança alimentando testes de cobertura, SAST e conformidade LGPD.
  - Agente Publicador gerando URL de prévia ao vivo e plano de rollback blue-green instantâneo.
  - Isolamento de memória viva por projeto (`geev.project_memory`) sem cruzamento de dados entre clientes.
- **Fase 10: Inteligência de Negócio & Placar da Fábrica:**
  - Motor de Preço Inteligente (`POST /v1/factory/estimate`) com cálculo por tipo de produto, complexidade e margem alvo.
  - Placar da Fábrica (`GET /v1/factory/scoreboard`) exibindo Lead Time médio, taxa de aprovação de primeira (92.5%), custos de IA/infra e tempo de decisão do Dono.
  - Integração no frontend em `CardDetailSheet` (botão Construir e badge de pipeline) e `FinancasTab` (cards executivos do placar e playbooks).

---

## [Fase 5] - Contratos dos Agentes — 2026-10-18

### Adicionado
- **Endpoints de Agente com ExecutionGuard (F5-01, F5-02):**
  - Endpoints `/v1/agent/cards`, `/v1/agent/cards/{id}`, `/v1/agent/cards/{id}/progress` e `/v1/agent/cards/{id}/submit-delivery`.
  - Resposta padronizada `423 Locked` imediata via `ExecutionGuard` se o agente estiver desligado (`active = false`) ou se o sistema estiver pausado pelo dono (`system.pausado = true`).
- **Autenticação Criptográfica de Agentes e Limite de Gasto (F5-03):**
  - Tokens de agentes validados via hash SHA-256 no banco de dados na tabela `geev.agents`.
  - Controle de teto de gastos (`spend_limit_cents`) e débito proporcional por progresso registrado.
- **Invariantes Invioláveis Testadas com -race (F5-04):**
  - Agente NUNCA consegue aprovar portões nem finalizar cartões. Ao submeter entrega, o cartão transita exclusivamente para `entrega_aprovar` (Portão 3) para revisão obrigatória do dono com ciência da Ficha de Risco.
  - Testes automatizados em `phase5_test.go` cobrindo bloqueio com 423, débito de gasto e avanço estrito para o portão de entrega.

---

## [Fase 4] - Fluxo Comercial — 2026-10-18

### Adicionado
- **Formulário Público de Demandas (Intake) com Honeypot e Auto-Retry (F4-01):**
  - Endpoint `POST /v1/intake` aberto com descarte silencioso de spam via honeypot (`website`).
  - Hashing de IP do solicitante para auditoria sem retenção de dados pessoais sensíveis (LGPD).
  - Tela pública `/intake` em Next.js com auto-retry progressivo para tolerar hibernação (cold start) da camada gratuita do Render.
- **Propostas Comerciais com Token Hash e Link Público de Aceite (F4-02, F4-03):**
  - Dono cria propostas com token criptográfico seguro de 32 bytes gerado e persistido apenas como hash SHA-256 no banco.
  - Página pública `/aceite` com visualização limpa do projeto, valores formatados em BRL e botão de aprovação pelo cliente.
  - Ao aceitar a proposta comercial, o cartão avança imediatamente para o estágio `prompt_aprovar` (Portão 2) com notificação em tempo real via SSE.
- **Conformidade LGPD para Clientes (F4-04):**
  - Endpoint `GET /v1/clients/{id}/export` para extração completa em JSON dos dados do titular (cartões, eventos e transações).
  - Endpoint `DELETE /v1/clients/{id}/purge` com expurgo dos dados a pedido do titular e registro imutável em `audit_log`.

---

## [Fase 3] - Comando e Risco — 2026-10-18

### Adicionado
- **Parser Determinístico de Comandos em Português (F3-01):**
  - Normalização semântica de acentos e casing para comandos: `aprovar <alvo>`, `recusar <alvo>`, `status`, `quanto gastei`, `lucro`, `parar sistema`, `retomar`.
  - Desambiguação estrita: se houver mais de um cartão para o mesmo cliente, retorna `ErrAmbiguousTarget` em vez de selecionar arbitrariamente.
- **Confirmação Criptográfica de Ações Destrutivas (F3-02):**
  - Ações como `recusar cartão` e `parar sistema` exigem confirmação explícita com token criptográfico de uso único e expiração (5 min).
  - Modal interativo na UI mobile exibindo o token e solicitando confirmação do dono.
- **Barra de Comando Global com Voz pt-BR (F3-03):**
  - Integração com Web Speech API com reconhecimento nativo no celular e execução transparente.
- **Ficha de Risco e Ciência Obrigatória (F3-04):**
  - Avaliação de cobertura de testes, varredura de segredos, dependências desatualizadas e pontos de atenção LGPD.
  - Bloqueio da aprovação de entrega na Bottom Sheet até que o dono marque a caixa de ciência da Ficha de Risco.

---

## [Fase 2] - Painel — 2026-10-18

### Adicionado
- **Shell da PWA e Navegação Mobile-First (F2-01):**
  - Barra de navegação inferior fixa com as 5 abas (`Hoje`, `Quadro`, `Meus apps`, `Finanças`, `Conexões`) e ícones SVG fiéis ao protótipo.
  - Cabeçalho dinâmico com título da aba, pílula de "Dinheiro parado em jogo", alternador de tema Claro/Escuro e botão SOS Parar/Retomar.
  - Faixa amarela de aviso quando o sistema está pausado com botão de retorno imediato.
  - Barra de comando flutuante com entrada de texto e microfone nativo (Web Speech API pt-BR) para comandos de voz.

- **Aba Hoje — 3 Decisões Prioritárias (F2-02):**
  - Ordenação automática dos cartões pendentes de decisão pelo maior valor em jogo.
  - Recomendações inteligentes de 1 linha por tipo de portão e botões para aprovação direta e abertura de detalhes.

- **Aba Quadro Kanban e Bottom Sheet de Cartão (F2-03):**
  - Colunas completas dos estágios de fluxo com scroll horizontal snap-x mobile.
  - Destaque em amarelo âmbar nas colunas que configuram portões do dono.
  - Bottom sheet animada com valores, custos, margem prevista (verde/vermelha), ficha de risco de entrega, histórico cronológico de eventos e botões de decisão (Aprovar, Pedir mais informações, Recusar e Fechar).

- **Aba Meus Apps (F2-04):**
  - Visualização de todos os apps da software house com cliente, hospedagem, banco, custo por mês, receita por mês e margem líquida.
  - Destaque automático de alerta para apps que "custam mais do que rendem" (candidatos a migração ou encerramento).
  - Modal para cadastro rápido de novo app.

- **Aba Finanças (F2-05):**
  - KPIs destacados de Receita Total, Custo Total e Lucro Líquido.
  - Barras gráficas de proporção visual (verde para receita e vermelha para custo) por app.
  - Custos agrupados por serviço (Render, Neon, Vercel, VPS...) e sugestão inteligente de consolidação.
  - Modal para lançamento manual de receitas ou despesas com persistência no banco.

- **Aba Conexões (F2-06):**
  - Lista fixa de 9 conectores com interruptor visual e política estrita de segurança (todos desligados por padrão).
  - Bloqueio e toast explicativo caso tente ligar conector sem configuração de chave segura.

- **Botão Parar/Retomar com Confirmação e Auditoria (F2-07):**
  - Modal com aviso das implicações antes de pausar agentes e conectores de emergência.
  - Registro de auditoria imutável na tabela `geev.audit_log`.

- **Tempo Real com Server-Sent Events (F2-08):**
  - Endpoint `GET /v1/stream` com broadcast in-memory e suporte a token via header ou query.
  - Hook `useSSE` no React invalidando automaticamente as consultas de dinheiro parado, cartões, apps e finanças sem recarregar a página.

- **PWA Instalável e Offline (F2-09):**
  - `manifest.json`, ícone SVG e Service Worker `sw.js` com cache de assets estáticos e isolamento de rotas de API.

- **Teste de Fluxo Ponta a Ponta (F2-10):**
  - Teste automatizado `TestFlow_AppsAndFinances` em Go que cadastra app, lança receita e despesa, calcula margens e consolida o lucro em menos de 100 ms.

---

## [Fase 1] - Fundação — 2026-10-18

### Adicionado
- **Documentação de Engenharia e Regras de Governança:**
  - `docs/PRD.md`: Especificação completa do produto baseada em `sprit01` com regras de portão, agente desligado por padrão e money-waiting.
  - `docs/ARCHITECTURE.md`: Arquitetura em camadas (Clean/Hexagonal/SOLID), decisões arquiteturais (ADR-001 a ADR-005) e justificativas de dependências mínimas.
  - `docs/SYSTEM_DESIGN.md`: Modelo relacional com 13 tabelas no schema `geev`, fluxo de autenticação com TOTP MFA AAL2 e topologia de implantação.
  - `docs/TASKS.md`: Rastreamento granular de tarefas divididas em 6 fases.
  - `docs/MEMORY.md`: Estado atual, decisões acordadas com o dono e riscos mitigados.
  - `.agents/rules/`: Regras mandatórias de workflow, arquitetura SOLID e segurança/LGPD para agentes autônomos.

- **Banco de Dados & Supabase Local (`/db`):**
  - Configuração do Supabase CLI com portas isoladas (`54332` Postgres, `54331` API Kong) e cadastro público desativado.
  - Migration `00001_schema.sql` (Goose): Criação do schema `geev`, enums, 13 tabelas relacionais, índices e triggers automáticos de `updated_at`.
  - Migration `00002_security.sql` (Goose): RLS habilitada e forçada em todas as tabelas, funções de segurança `is_owner()` e `can_execute()`, triggers de imutabilidade (bloqueio de `UPDATE`/`DELETE` em `approvals`, `card_events` e `audit_log`) e policies para o dono com AAL2.
  - Suite de testes automatizados pgTAP em `db/supabase/tests/security_test.sql` cobrindo o bloqueio RLS para anônimos e integridade de schema (100% aprovada).

- **Backend Go (`/api`):**
  - Domínio puro: tipos monetários seguros em centavos `Money` (`BRL`), máquina de estados de cartões `CardStatus` com portões de aprovação obrigatórios, decisão de portão `ApprovalDecision`, principal tipado `OwnerPrincipal` vs `AgentPrincipal`, guarda de execução `ExecutionGuard` e abstração de relógio `Clock`.
  - Casos de uso (`/internal/app`): `DecideGateUseCase`, `PauseSystemUseCase`, `ResumeSystemUseCase`, `GetMoneyWaitingUseCase` com interfaces de portas granulares (`CardRepository`, `ApprovalRepository`, `AuditRepository`, `SystemRepository`).
  - Adaptadores Postgres (`/internal/adapters/postgres`): Queries SQL tipadas compiladas via `sqlc`, repositórios com contexto transacional seguro e Unit of Work (`UnitOfWork`).
  - Adaptador de Autenticação (`/internal/adapters/auth`): Validador de JWT Supabase exigindo `sub == OWNER_USER_ID` e nível de autenticação `aal == aal2`.
  - Middlewares HTTP: Injeção de `X-Request-ID`, structured logging JSON via `log/slog` com anonimização de PII, recoverer e autenticação JWT.
  - Handlers HTTP stdlib: Liveness `/healthz`, readiness `/readyz`, `POST /v1/cards/{id}/decision`, `POST /v1/system/pause`, `POST /v1/system/resume` e `GET /v1/money-waiting`.
  - Testes unitários com fakes de portas e testes de integração com banco de dados, todos validados com `-race`.

- **Contratos & Contêiner:**
  - Especificação OpenAPI 3.0 completa em `api/openapi.yaml`.
  - Dockerfile multi-stage com compilação estática Go e base `gcr.io/distroless/static-debian12:nonroot`, gerando imagem segura com apenas 23.7 MB.

- **Frontend Mobile-First (`/web`):**
  - Aplicação Next.js 16 (App Router) configurada com `output: 'export'` para deploy estático no Cloudflare Pages.
  - Tailwind CSS v4 e paleta de cores fieis ao protótipo aprovado (`docs/prototype/central-prototipo.html`), com suporte nativo a temas Claro e Escuro via CSS variables.
  - Provedor TanStack Query com revalidação automática e API client centralizado (`shared/api-client`) com injeção automática de JWT e tratamento de erros.
  - Tela de login com fluxo de 2FA TOTP (Supabase Auth).
  - Shell do dashboard mobile-first (360px) com cabeçalho de status da operação, visualizador de "Dinheiro parado" em tempo real, botão de emergência Parar/Retomar sistema e barra de comando.

- **CI & DevOps:**
  - Pipeline GitHub Actions em `.github/workflows/ci.yml` cobrindo Go format, vet, race tests, Docker build, compilação de static export do Next.js, migrations goose e testes pgTAP do Supabase.
