# Memory — GeevheitNeural

> Memória viva do projeto. Atualize a cada decisão, descoberta ou fim de fase.
> Agentes de código: **leia este arquivo antes de começar qualquer tarefa.**

## Estado atual
- **Sprint:** 03 — Expansão Comercial e Fábrica Autônoma (Jarvis Ativo).
- **Status do Projeto:** Jarvis, Fábrica de IA e Captação de Pedidos **100% OPERACIONAIS EM PRODUÇÃO**.
- **Marcos Recentes:**
  - Domínio oficial de produção ativo: `https://geevheit.vercel.app`.
  - Chat Interativo do Jarvis com alternador de voz (🔊/🔇) e histórico persistente.
  - Conexão do Google Gemini Flash para esteira autônoma com custo R$ 0,00 de tokens.
  - Criação da Landing Page de alta conversão do produto real: TransportSim 360 (`https://geevheit.vercel.app/transportsim`) com simulador rodoviário e modal de matrícula conectado à API `/v1/intake`.
  - Integração oficial do site da Software House (`https://geevheith-site.vercel.app`) para gravar orçamentos e novos projetos diretamente no banco Supabase.
  - **Exclusão Definitiva e Limpeza de Dados Legados:** Adicionada funcionalidade para deletar demandas/cartões permanentemente (via botão no detalhe do cartão, comandos de voz/texto "excluir/deletar [cliente]" e rota `DELETE /v1/cards/{id}`), botão de zerar dados legados no Hub, além de correção do fallback local e persistência imediata no `localStorage` para evitar que itens excluídos ou recusados voltem ao recarregar a tela.
  - **Homologação 100% Ponta a Ponta do `sprit02` (Fases 6 a 10) e Prontidão para Deploy (2026-10-29):**
    - Comando `"construir [projeto/cliente]"` integrado no parser Go (`IntentBuild`), `ProcessCommandUseCase`, `CommandLine` e `page.tsx`.
    - Pipeline de 7 agentes (`Requisitos → Arquiteto → Dev → QA → Segurança → Publicador → Suporte`) com roteamento de modelos (`economy/fast` vs `critical/reasoning`), teto de gasto (`spend_limit_cents`), descarte de sandbox na pausa de emergência e marcação automática `revisão_humana_obrigatória` em domínios sensíveis (pagamentos, auth, LGPD).
    - `CardDetailSheet` completa com Especificação Executável BDD (`/escopo?token=...`), Prévia ao Vivo (`/preview`), visualizador de código do workspace, Plano de Reversão Automático (`POST /v1/cards/{id}/rollback`), Memória Viva do Projeto (`GET/POST /v1/cards/{id}/memory`) e Ficha de Risco automática compatível com formato de tupla e objeto.
    - `AgentsCockpit` ("Fábrica de IA") com os 7 papéis, chave Ligar/Desligar individual (`POST /v1/agents/{id}/toggle`), Placar da Fábrica (`scoreboard`), Biblioteca de Playbooks e Módulos Reutilizáveis com Nível de Autonomia (1 a 3) e promoção de padrões (`POST /v1/modules`), Suíte de Evals (`POST /v1/factory/evals/run`) e Monitoramento Pós-Entrega do Agente de Suporte (`POST /v1/apps/{id}/monitor`).
    - Motor de Preço, Prazo e Margem Assistido integrado ao modal `NewCardModal`.
  - **Ativação Real de Ordens via Jarvis / Jake & Endurecimento de Segurança (2026-10-29):**
    - Identificada e corrigida a causa raiz de ordens do Jarvis/Jake não entrarem em produção: a gaveta de chat (`JarvisChatDrawer.tsx`) apenas retornava strings locais de simulação e não chamava a API da central.
    - Conectado `JarvisChatDrawer` ao `onExecuteCommand` (`api.processCommand`), com suporte a síntese de voz (TTS) do resultado real e invalidação em tempo real do cache do Kanban e métricas.
    - Adicionado suporte a prefixos de assistente no parser determinístico Go (`parser.go`), permitindo frases como *"Jarvis, criar projeto X"* ou *"Jake, construir sistema Y"*.
    - Criada a intenção `IntentCreate` para registro automático de demandas por comando de voz/texto.
    - O comando `IntentBuild` agora cria a demanda automaticamente sob demanda se ela ainda não existir, iniciando a esteira dos 7 agentes de IA na mesma hora.
    - **Endurecimento de Segurança no Aceite de Escopo:** Eliminada a previsibilidade do link de escopo gerando tokens aleatórios criptograficamente seguros (32 bytes `sc_...`) salvos apenas como hash SHA-256 no Postgres.
    - Acoplamento do sinal de entrada à liquidação por PIX via `ProcessPixWebhookUseCase`.
    - Remoção de fallback local silencioso em `/escopo`, garantindo resposta de erro explícita caso o link seja inválido.
    - **Solução Definitiva para Retorno de Dados Excluídos:** Corrigida a query `DeleteCard` do PostgreSQL para habilitar `set_config('geev.allow_card_delete', 'on', true)` no cascade, e adicionada lista negra imutável de IDs no frontend (`geev_deleted_card_ids` / `geev_deleted_app_ids`) via `useMemo`. Quando o dono clica em excluir ou em "Limpar Dados Legados", os itens são eliminados e nunca mais voltam ao atualizar a página.
- **Qualidade e Conformidade:**
  - Backend Go 100% verde (`go test -race ./...`).
  - Frontend Next.js estático 100% verde (`npm run build` / `out/`).
  - Banco de Dados Supabase Nuvem com RLS forçada e Goose v11 (`00011_allow_card_deletion.sql`).
- **Última atualização:** 2026-10-29.

## Respostas do dono (2026-10-18)
1. **Protótipo:** salvo em [`docs/prototype/central-prototipo.html`](./prototype/central-prototipo.html). É a referência de telas, textos e tokens de cor (claro e escuro).
2. **Tailwind:** v4.
3. **Host da API:** Render gratuito na Fase 1. Formulário público com **retry automático** por causa da hibernação; alternativa documentada.
4. **Supabase:** só local (Supabase CLI) até a Fase 2. Na nuvem: projeto **novo e separado, região São Paulo**.
5. **LGPD:** isolamento por cliente no mesmo banco, com exportação e exclusão por cliente.
6. **Pedir mais informações:** na entrega volta para *Em execução*; nos outros estágios fica onde está e registra o pedido.
7. **Recusar:** estado final `recusado`, fora do Kanban, visível no histórico, **restaurável pelo dono**, com `audit_log`.
8. **Moeda:** só BRL; valores em centavos (inteiro) + coluna `currency`.
9. **Tema:** claro e escuro (automático pelo sistema + alternância manual).

## O que o protótipo define (e como foi traduzido)
- Abas: **Hoje, Quadro, Meus apps, Finanças, Conexões** (navegação inferior com ícones SVG).
- Cabeçalho: título da aba + pílula "R$ X parados esperando você · N decisões" + botão vermelho **Parar/Retomar**.
- Barra de comando fixa acima da navegação: "Dite uma ordem. Ex.: aprovar Casa Verde", botões **Falar** e **Enviar**.
- Detalhe do cartão em *bottom sheet*: valor cobrado, custo estimado, margem, ficha de risco, histórico, ações.
- Rótulos de ação por estágio (ex.: "Aprovar prompt e liberar execução") — viram textos da UI.
- **Pausado** bloqueia liberar execução (aprovar *Prompt* e avançar *Em execução*). Regra levada ao domínio (`card.Transition.TouchesExecution`).
- No protótipo, *Entrada* também entra em "Hoje" e no contador. **Decisão:** "Hoje" considera portões + Entrada (como no protótipo); o **contador "Dinheiro parado"** soma **só portões** (como no spec). Ver pergunta 1 abaixo.
- Fonte: Instrument Sans. Cores: `--acc #0e6b73` (claro) / `#4fc3cc` (escuro); paleta completa em `web/src/app/globals.css`.

## Contexto em uma frase
PWA mobile-first em que **um único dono** opera uma software house: decide, aprova e dá ordens; agentes de IA existem só como contrato, **desligados**.

## Glossário
| Termo | Significado |
|---|---|
| Portão | Estágio que só avança por decisão explícita do dono: *Proposta para aprovar*, *Prompt para aprovar*, *Entrega para aprovar*. |
| Dinheiro parado | Soma de `price_cents` dos cartões em portões. |
| Parar | Pausa de emergência: bloqueia toda execução de agentes e conectores. |
| ExecutionGuard | Peça única do domínio que decide se uma execução pode acontecer. |
| Ficha de risco | Testes, segredos, dependências e LGPD mostrados antes de aprovar uma entrega. |

## Decisões tomadas (resumo — detalhes em ARCHITECTURE §6)
- Go com `net/http` da stdlib (sem chi), sqlc + pgx, goose.
- API é o **único** ponto de ação; o front usa Supabase só para auth.
- RLS vale também para o tráfego da API (role + claims por transação).
- Dinheiro em centavos (`bigint`).
- Tempo real por SSE com broker em memória (trocável).
- Next.js em static export ⇒ Cloudflare Pages.
- Parser de comandos determinístico (sem LLM).

## Convenções
- Código e identificadores em **inglês**; textos de UI, docs e mensagens de commit em **português**.
- Commits: Conventional Commits (`feat(api): …`, `fix(web): …`).
- Migrations: `db/migrations/NNNNN_descricao.sql`, sempre com `-- +goose Up` e `-- +goose Down`.
- Erros da API: `{"error": {"code": "...", "message": "..."}}`; códigos estáveis (ex.: `gate_requires_owner`, `system_paused`, `ambiguous_target`).

## Armadilhas conhecidas e resolvidas
- **Conflito de portas do Docker Supabase:** portas `54321` e `54322` padrão estavam alocadas por outro container da máquina. Ajustadas em `config.toml`, `.env.example` e `Makefile` para `54331` (Kong API) e `54332` (Postgres).
- **PostgREST Schema Cache:** listar `geev` em `config.toml` antes do goose rodar gera erro `3F000` de inicialização no PostgREST. Como a API Go é o ponto único de entrada para `geev`, PostgREST foi configurado para schemas públicos.
- **Goose Block Delimiters:** blocos e triggers PL/pgSQL com `$$` exigem `-- +goose StatementBegin` e `-- +goose StatementEnd`.
- Pooler do Supabase em modo transação (6543) **não** aceita prepared statements nomeados ⇒ pgx com `QueryExecModeSimpleProtocol`/`Exec`.
- `LISTEN/NOTIFY` não funciona pelo pooler em modo transação; por isso o tempo real usa broker em memória (uma instância).
- Next.js static export não suporta rotas dinâmicas sem `generateStaticParams` ⇒ usar query string/drawer.
- O protótipo no claude.ai precisa de JavaScript/login; a cópia local em `docs/prototype/` é a referência.

## Perguntas em aberto
1. **Contador "Dinheiro parado":** o spec diz "cartões parados em portões do dono"; o protótipo soma também a *Entrada*. Implementado como no spec (só portões). Confirmar.

## Histórico de fases
| Fase | Status | Resumo |
|---|---|---|
| 0 | Aprovada (2026-10-18) | Documentação de produto, arquitetura, design, regras e tarefas. |
| 1 | Concluída (2026-10-18) | Fundação: banco com RLS + pgTAP, API Go Clean Arch com -race, Docker distroless < 24MB, Next.js static export com Tailwind 4 e 2FA TOTP, CI GitHub Actions, README e CHANGELOG. |
| 2 | Concluída (2026-10-18) | Painel completo: 5 abas (Hoje, Quadro, Meus apps, Finanças, Conexões), Bottom Sheet de cartão, SSE em tempo real, barra de comando com voz, PWA offline e teste de fluxo e2e em 80ms. |
| 3 a 5 | Concluída (2026-10-19) | Fechamento do Sprint 01: PWA offline com background sync, rotas públicas (/intake, /escopo, /aceite), health check e métricas. |
| 6 a 10 | Concluída (2026-10-19) | Fábrica Autônoma de Software (Sprint 02): pipelines multi-agente, playbooks, sandbox de preview, placar de margem/lead-time, estimativa inteligente de preço. |
| Deploy Prep | Concluída (2026-10-19) | Preparação completa para produção: render.yaml (IaC), docs/DEPLOYMENT.md, scripts/pre_deploy_check.sh, CodeGraph indexado. |
| Deploy Produção | Concluída (2026-10-21) | Banco Supabase migrado (Goose v6 com RLS forçada), repositório GitHub sincronizado, usuário Dono provisionado no Auth, Frontend Vercel LIVE em produção (https://web-lake-eight-gwn3vopv94.vercel.app). |
| Etapa 0 (spec01) | Concluída (2026-10-21) | Hub de Quadros (Tela Inicial com 7 etapas, grade/lista, saúde infra) + Logout Seguro com 2FA AAL2 (branch feature/home-hub-logout). |
| 13 (spec02) | Concluída (2026-10-21) | Portal Exclusivo do Cliente (/portal): token criptografado SHA-256, isolamento estrito de dados internos, polling 15s, aceite formal imutável em audit_log (branch feature/fase-13-portal-cliente). |
| 14 (spec02) | Concluída (2026-10-23) | Gestão de Mensalidades (MRR) e Recorrência: modelo de assinaturas e faturas, cálculo de MRR sem float, rotina diária idempotente e painel com alertas em Finanças (branch feature/fase-14-mrr). |
| 15 (spec02) | Concluída (2026-10-23) | Notificações Operacionais & Webhooks de Saída: mensageria com MultiNotifier (Telegram, WhatsApp, E-mail), webhooks com assinatura HMAC-SHA256 e proteção anti-SSRF, histórico imutável de entregas, gestão de preferências na aba Conexões e especificação detalhada do Grafo de Agentes em docs/WEBHOOKS_AND_ORCHESTRATION.md (branch feature/fase-15-notifications-webhooks). |
| 16 (prompt.md) | Concluída (2026-10-23) | Fábrica Autônoma 100% Gratuita: GeminiAgentRunner (Free Tier 1.500 req/dia e rate limiter de 15 RPM), LocalSandboxManager (isolamento de diretório efêmero e anti-path traversal), LocalPublisher (versionamento Git local atômico) e rotas HTTP de prévia ao vivo (/v1/cards/{id}/preview) sem custos adicionais (branch feature/fase-16-autonomous-factory). |
| 17 (Melhorias) | Concluída (2026-10-27) | Motor Groq de Alta Velocidade (GroqAgentRunner em Go com gpt-oss-120b/qwen3.8-27b, 0,19s por inferência), Assistente Jarvis/Jake com alternância de voz masculina/feminina, tratamento exclusivo "Senhor", painel de voz no drawer e política estrita de blindagem a custo zero (Zero Risk Billing). |




