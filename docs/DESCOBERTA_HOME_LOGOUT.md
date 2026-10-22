# Descoberta: Tela Inicial (Hub de Quadros) + Botão de Sair (Fase 0)

> Documento produzido em conformidade estrita com a Seção 1 da `spec01.md`.
> **Data:** 2026-10-21  
> **Branch de trabalho:** `feature/home-hub-logout`  
> **Status:** Concluído (Fase 0 - Linha de Base Aprovada)

---

## 1. Stack Tecnológica

| Componente | Tecnologia | Versão / Detalhes |
|---|---|---|
| **Frontend** | Next.js (App Router, Turbopack) | `16.4.2`, React 19, TypeScript 5 |
| **Exportação Front** | Estático (`output: 'export'`) | PWA sem SSR / sem Server Actions |
| **Estilização** | Tailwind CSS v4 + Vanilla CSS | Design tokens via CSS variables em `globals.css` |
| **Gerenciador de Estado / API** | TanStack React Query | v5 (`useQuery`, `useMutation`, `QueryClient`) |
| **Backend API** | Go | 1.22+, `net/http` stdlib, Clean Architecture |
| **Banco de Dados** | PostgreSQL (Supabase Nuvem / Local) | Migrações Goose v7, RLS forçada em 100% das tabelas |
| **Autenticação** | Supabase Auth (GoTrue) | JWT com TOTP 2FA obrigatório (AAL2) |
| **Comunicação Tempo Real** | Server-Sent Events (SSE) | `/v1/events` com broker em memória no Go |
| **Gerenciadores de Pacote** | `npm` (web) e `go modules` (api) | Monorepo estruturado em `/api`, `/web`, `/db`, `/docs` |

---

## 2. Tecnologia do Canvas e Visualização

- **Situação no repositório:** O projeto atual GeevheitNeural **não** possui biblioteca de canvas infinito de nós (ex: React Flow, tldraw, Konva, Excalidraw).
- **Interface Real:** O sistema implementa uma interface operacional mobile-first com abas especializadas:
  1. `HojeTab`: Painel de decisões do dia ordenadas por dinheiro em jogo.
  2. `QuadroTab`: Kanban responsivo (`snap-x snap-mandatory`) com as 7 etapas do ciclo de vida dos cartões.
  3. `MeusAppsTab`: Inventário de quadros/aplicações por cliente, custos, receitas e margem.
  4. `FinancasTab`: DRE e lançamentos financeiros por serviço/cliente.
  5. `ConexoesTab`: Conectores e integrações (somente leitura / desligados por padrão).
- **Regra de preservação:** Em estrito cumprimento à regra *"NÃO QUEBRAR NADA DO QUE JÁ FUNCIONA"*, a experiência de navegação, visualização Kanban e o fluxo operacional móvel serão totalmente preservados, adicionando o Hub consolidado e a gestão de sessão.

---

## 3. Modelo de Dados

### Tabelas Principais (Schema `geev`):
1. **`geev.cards`**:
   - `id` (UUID), `client_id` (UUID), `app_name` (text), `title` (text), `type` (orçamento, bug, suporte, financeiro).
   - `price_cents` (bigint), `est_cost_cents` (bigint), `currency` (BRL).
   - `stage` (entrada, proposta_aprovar, aguardando_cliente, prompt_aprovar, em_execucao, entrega_aprovar, entregue, recusado).
   - `priority` (int), `created_at`, `updated_at`.
2. **`geev.apps`** (Quadros/Projetos da empresa):
   - `id`, `client_name`, `name`, `hosting`, `database`, `repository`, `domain`, `status`, `monthly_cost_cents`, `monthly_revenue_cents`, `net_margin_cents`.
3. **Imutabilidade e Auditoria**:
   - `geev.card_events`: histórico imutável de eventos de cada cartão.
   - `geev.approvals`: histórico imutável de aprovações exclusivas do dono nos portões.
   - `geev.audit_log`: trilha de auditoria para segurança e LGPD.
4. **Camada de Apresentação (`web`)**:
   - Mapeada via interfaces TypeScript em `web/src/shared/api-client/endpoints.ts` (`CardItem`, `AppItem`, `FinancesOverview`, `ConnectorItem`, `CardDetails`).
   - Caches gerenciados pelo TanStack Query com fallbacks fiéis ao protótipo aprovado caso a API esteja operando offline.

---

## 4. Etiquetas de Produção

As etapas de produção refletem o ciclo de vida dos cartões no sistema:

| Etapa / Etiqueta | Nome na UI | Portão do Dono? | Cor / Destaque |
|---|---|:---:|---|
| `entrada` | Entrada | Não | Neutro / Linha |
| `proposta_aprovar` | Proposta para aprovar | **Sim (🛡️)** | `--warn` (`#b45309`) / Fundo `--warns` |
| `aguardando_cliente` | Aguardando cliente | Não | Neutro |
| `prompt_aprovar` | Prompt para aprovar | **Sim (🛡️)** | `--warn` (`#b45309`) / Fundo `--warns` |
| `em_execucao` | Em execução | Não | `--acc` (`#0e6b73`) |
| `entrega_aprovar` | Entrega para aprovar | **Sim (🛡️)** | `--warn` (`#b45309`) / Fundo `--warns` |
| `entregue` | Entregue | Não | `--ok` (`#027a48`) |

Além dos estágios, os cartões recebem etiquetas de **Tipo de Demanda**:
- **Orçamento**: `#0e6b73` (Verde-azulado)
- **Bug**: `#b42318` (Vermelho)
- **Suporte**: `#b45309` (Âmbar)
- **Financeiro**: `#6d4bb5` (Roxo)

---

## 5. Salvamento do Estado e Sincronização

- Não há rascunhos voláteis não salvos no cliente; toda alteração (aprovação, transição, cadastro de app, novo lançamento) executa mutação direta via TanStack Query para a API Go.
- O backend Go registra eventos em `geev.card_events` e `geev.audit_log` de forma atômica e transacional.
- Atualização em tempo real: realizada via Server-Sent Events (`useSSE`) para flags como pausa de emergência (`isPaused`).

---

## 6. Roteamento e Estrutura de Telas

- **`/`**: Rota principal autenticada (`web/src/app/page.tsx`).
  - Se não houver sessão ativa: renderiza `LoginForm` com fluxo 2FA TOTP.
  - Se autenticado: renderiza o Dashboard com navegação entre abas (`Hoje`, `Quadro`, `Meus apps`, `Finanças`, `Conexões`), Header com dinheiro parado e CommandLine.
- **Rotas Públicas (Static Export)**:
  - `/intake`: Formulário público de novos pedidos.
  - `/escopo`: Apresentação de propostas comerciais.
  - `/aceite`: Assinatura e aceite digital com validação de token.

---

## 7. Autenticação e Gestão de Sessão Atual

- Cliente Supabase instanciado em `web/src/shared/auth/supabase.ts` com `persistSession: true` e `autoRefreshToken: true`.
- Validação no Go: middleware `RequireOwnerAuth` valida claims e exige `sub == OWNER_USER_ID` e `aal == aal2`.
- **Gargalo / Oportunidade identificada**:
  - A interface gráfica atual no Header **não possui menu de usuário nem botão de Logout**.
  - O logout deve ser implementado no Header (com avatar/iniciais do dono, dropdown seguro, modal de confirmação, encerramento de sessão via `supabase.auth.signOut()`, cancelamento de timers e limpeza do cache do TanStack Query).

---

## 8. Sistema de UI e Acessibilidade

- Design System com CSS variables no `:root` e `[data-theme="dark"]` (`globals.css`).
- Tipografia: Instrument Sans / sans-serif do sistema.
- Suporte a tema claro e escuro persistido em `localStorage.getItem("geev_theme")`.
- Resolução de referência: Mobile-first a partir de 360 px, com adaptação fluida para tablet e desktop.
- Elementos com IDs únicos e semânticos para testes e automação.

---

## 9. Linha de Base de Testes e Build

Executados antes de qualquer modificação de código:

1. **Frontend (`/web`)**:
   - Comando: `npm run build`
   - Resultado: **100% Sucesso**
   - Saída: `✓ Compiled successfully in 1217ms`, todas as páginas estáticas (`/`, `/_not-found`, `/aceite`, `/escopo`, `/intake`) geradas sem erro.
2. **Backend API (`/api`)**:
   - Comando: `go test -race ./...`
   - Resultado: **100% Sucesso** (todos os pacotes `ok`).
   - Comando: `go build -o /dev/null ./cmd/geevheit-api`
   - Resultado: **100% Sucesso** (compilação limpa).

---

## 10. Matriz de Riscos

| Risco | Impacto | Mitigação |
|---|---|---|
| Quebra de exportação estática | Falha no deploy da Vercel / Cloudflare | Manter `output: 'export'` e testar `npm run build` após cada alteração |
| Perda de estado de sessão ao sair | Sessão travada ou token residual | Garantir `supabase.auth.signOut()`, `queryClient.clear()` e redirecionamento limpo |
| Regressão na visualização mobile | Quebra de layout em 360 px | Testar responsividade e manter BottomNav + Header perfeitamente alinhados |
| Incompatibilidade com contratos da API | Erro 401/403/500 | Utilizar apenas chamadas compatíveis e schemas já existentes |
