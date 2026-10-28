# Changelog — GeevheitNeural

Todas as mudanças notáveis deste projeto serão documentadas neste arquivo seguindo o formato [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/) e [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Sprint 03 - Melhoria] - Exclusão Definitiva de Demandas e Limpeza de Dados Legados — 2026-10-28

### Adicionado
- **Exclusão Definitiva de Cartões/Demandas (`DELETE /v1/cards/{id}`):**
  - Caso de uso `DeleteCardUseCase` e handler HTTP `HandleDeleteCard` no backend Go.
  - Método `Delete(ctx, id)` adicionado na porta `CardRepository` e repositório Postgres.
  - Migration Goose `00011_allow_card_deletion.sql` para permitir remoção em cascata controlada pelo dono mantendo a imutabilidade estrita do `audit_log`.
  - Suporte a comandos de texto e voz com intenção de exclusão (`excluir`, `deletar`, `remover`, `apagar`) com token de confirmação de uso único.
- **Interface e Experiência do Dono:**
  - Botão **"🗑️ Excluir Demanda Permanentemente"** com confirmação de segurança dentro da gaveta de detalhe do cartão (`CardDetailSheet.tsx`).
  - Botão **"🧹 Limpar Dados Legados"** no cabeçalho do Hub (`HomeHubTab.tsx`) para zerar dados legados/demonstração com 1 clique e confirmação.

### Corrigido
- **Persistência de Decisões e Prevenção de Retorno de Dados Legados:**
  - Salvamento imediato de qualquer decisão (aprovação ou recusa) no `localStorage` (`geev_local_cards`).
  - Flag `geev_cleared_legacy = true` para respeitar painel zerado sem forçar re-injeção de `REAL_CARDS` ou `DEMO_CARDS` ao atualizar a página.
  - Ajuste de fallback para priorizar dados retornados da API mesmo quando a lista estiver vazia (`[]`).

---

## [Sprint 03 - Fase 17] - Motor Groq LPU, Assistente Jarvis/Jake e Blindagem Custo Zero — 2026-10-27

### Adicionado
- **Motor Groq de Inferência Ultrarrápida (`GroqAgentRunner`):**
  - Implementação completa em Go conectando à API Groq Cloud (chips LPU com tempo de resposta de ~0,19s).
  - Suporte ao modelo open-source de 120 bilhões de parâmetros (`openai/gpt-oss-120b`) e `qwen/qwen3.8-27b`.
  - Formatação estruturada em JSON (`response_format: {"type": "json_object"}`) e sanitização de markdown fences.
  - Suíte de testes unitários com mock HTTP e validação `-race` (`groq_runner_test.go`).
- **Assistente Operacional com Alternância de Personas (Jarvis & Jake):**
  - Seletor dinâmico de personas no cabeçalho do chat:
    - 🧠 **Jarvis (Masculino):** Tom grave (pitch 0.85), avatar teal e voz masculina.
    - ✨ **Jake (Feminina):** Tom suave (pitch 1.15), avatar violeta/rosa e voz feminina.
  - Painel de voz integrado para ajustes de velocidade (rate), tom (pitch), teste instantâneo de áudio e escolha de vozes instaladas no sistema operacional.
  - Formalização executiva: tratamento exclusivo e respeitoso como **"Senhor"** em todas as mensagens e respostas por voz.
- **Blindagem e Política de Custo Zero (Zero Risk Billing):**
  - Configuração estrita para uso exclusivo de provedores com planos Always Free ou sem cartão de crédito (Google Gemini Free Tier, Groq Free Tier, ViaCEP pública e Firebase Spark).

---

## [Sprint 03 - Fase 16] - Fábrica Autônoma de Software 100% Gratuita (Free Tier) — 2026-10-23

### Adicionado
- **Adapter de IA Google Gemini (`GeminiAgentRunner`):**
  - Implementação real de `pipeline.AgentRunner` conectando à API do Google Gemini (Google AI Studio Free Tier).
  - Rate Limiter determinístico interno com intervalo mínimo entre requisições respeitando estritamente o limite gratuito de 15 RPM (1.500 req/dia com R$ 0,00 de custo).
  - Sanitização de Markdown fences (`cleanJSONMarkdown`) e tradução determinística para artefatos de negócio por papel.
  - Testes de unidade completos com mock HTTP validando ausência de chave, sucesso, portão de entrega obrigatório e tratamento de erros de API.
- **Gerenciador de Sandboxes e Workspaces Efêmeros (`LocalSandboxManager`):**
  - Criação isolada de diretórios de trabalho por projeto em `scratch/builds/<card-id>/`.
  - Validação estrita de segurança anti-path traversal (bloqueando `..`, caminhos absolutos e escapes de diretório).
  - Escrita multi-file atômica de código-fonte (`WriteFiles`) e leitura recursiva de artefatos (`ReadFiles`).
  - Execução segura de comandos de validação/build em container local ou shell isolado (`RunValidation`).
  - Testes de unidade cobrindo ciclo de vida, integridade de arquivos, rejeição de travessia maliciosa e limpeza.
- **Conector de Versionamento Git e Deploy de Prévia (`LocalPublisher`):**
  - Inicialização automática de repositório Git local (`git init`), configuração de autor e commits atômicos padronizados por build.
  - Endpoints REST para visualização e inspeção de código gerado:
    - `GET /v1/cards/{id}/preview`: entrega a interface web compilada (`index.html`) para teste interativo no navegador.
    - `GET /v1/cards/{id}/workspace`: lista todos os arquivos e conteúdos do sistema gerado em JSON (protegido pelo Dono).
- **Integração no Composition Root e Invariantes:**
  - Injeção em `cmd/geevheit-api/main.go` com comutação via `AGENT_RUNNER_TYPE=gemini|mock` e detecção de `GEMINI_API_KEY`.
  - Conexão no caso de uso `StartBuildUseCase` para geração física de arquivos no `RoleDev` e publicação atômica no `RolePublicador`.
  - 100% dos testes da suíte passando com `-race` (`go test -race ./...`).

---

## [Sprint 03 - Fase 15] - Notificações de Negócio, Webhooks com HMAC/SSRF e Grafo de Agentes — 2026-10-23


### Adicionado
- **Notificações Operacionais Multicanal (`MultiNotifier`):**
  - Modelagem e migração `00010_notifications_and_webhooks.sql` criando `geev.notification_preferences`, `geev.webhooks` e `geev.webhook_deliveries` com RLS ativada e forçada.
  - Interface desacoplada `Notifier` e adaptador `MultiNotifier` com suporte a Telegram Bot, WhatsApp Gateway e E-mail operacional.
  - Disparo de alertas críticos: novos pedidos em `/intake`, confirmação de pagamento PIX, aceite formal no Portal e atraso de mensalidades.
  - Endpoints REST autenticados: `GET /v1/notifications/preferences` e `PUT /v1/notifications/preferences`.
- **Webhooks de Saída com Assinatura Criptográfica e Proteção Anti-SSRF:**
  - Assinatura HMAC-SHA256 (`X-Geevheit-Signature: sha256=...`) com carimbo de tempo Unix (`X-Geevheit-Timestamp`) para prevenção de ataques de replay.
  - Validador estrito anti-SSRF bloqueando ranges privados (RFC 1918), loopback (`127.0.0.1`, `localhost`) e metadados de nuvem (`169.254.169.254`).
  - Despachante assíncrono em segundo plano com timeout de 10s e registro imutável em `geev.webhook_deliveries` (status code, tempo de resposta e payload).
  - Endpoints REST autenticados: `POST /v1/webhooks`, `GET /v1/webhooks`, `DELETE /v1/webhooks/{id}`, `GET /v1/webhooks/deliveries` e `POST /v1/webhooks/{id}/redeliver`.
- **Interface na Aba Conexões:**
  - Seções integradas: "Notificações Operacionais do Dono" e "Webhooks de Saída (Eventos em Tempo Real)".
  - Gerenciador de preferências com toggles por canal e por tipo de evento crítico.
  - Formulário de cadastro de webhook com visualização do secret HMAC para configuração no cliente.
  - Histórico de entregas com status HTTP, duração em ms, carimbo de data e botão de reenvio manual imediato.
- **Documentação e Arquitetura do Grafo de Agentes (F15-03):**
  - Criação de `docs/WEBHOOKS_AND_ORCHESTRATION.md` com exemplos prontos de validação HMAC em Node.js/Express e workflows no n8n.
  - Especificação arquitetural completa do Grafo de Agentes determinístico (Requisitos → Arquiteto → Dev → QA → Deploy → Aceite do Cliente), com isolamento em sandbox, portões com 2FA do Dono e garantias de segurança via `ExecutionGuard`.

---

## [Sprint 03 - Fase 14] - Gestão de Mensalidades (MRR) e Recorrência — 2026-10-23

### Adicionado
- **Motor e Modelagem de Assinaturas Recorrentes:**
  - Migração `00009_subscriptions_and_mrr.sql` criando `geev.subscriptions` e `geev.subscription_charges` com RLS forçada e restrição de unicidade por ciclo.
  - Domínio `subscription` com cálculo de MRR estritamente em centavos inteiros (sem ponto flutuante), suportando periodicidades mensal, trimestral, semestral e anual.
  - Métricas de faturamento recorrente (MRR Atual, MRR Previsto, Ativas, Pausadas, Inadimplência e Churn).
  - Rotina diária idempotente (`POST /v1/subscriptions/process-daily`) para atualizar cobranças vencidas para atrasadas e gerar faturas futuras.
  - Endpoints REST autenticados: `POST /v1/subscriptions`, `GET /v1/subscriptions`, `PATCH /v1/subscriptions/{id}/status`, `GET /v1/finances/mrr`, `GET /v1/finances/charges`, `POST /v1/finances/charges/{id}/pay`.
  - Baixa de mensalidade integrando crédito de receita automático no Ledger financeiro e gravação de histórico imutável em `audit_log`.
- **Interface na Aba Finanças:**
  - Seletor de abas: "Mensalidades & MRR" e "Visão Geral & Lançamentos".
  - Cards de KPIs com MRR Atual, Previsto, Assinaturas Ativas e alerta de Inadimplência.
  - Painel de faturas vencidas e a vencer com botão de baixa direta ("Marcar Pago").
  - Tabela de assinaturas com badges de status e ações rápidas (Pausar, Reativar, Cancelar).
  - Modal de cadastro de nova assinatura recorrente com seleção de app, dia de vencimento e periodicidade.
- **Qualidade & Segurança:**
  - Testes unitários com corrida (`-race`) cobrindo cálculo exato de MRR, idempotência da rotina diária e liquidação de mensalidades.

---

## [Sprint 03 - Fase 13] - Portal Exclusivo do Cliente e Aceite Formal — 2026-10-21

### Adicionado
- **Portal Exclusivo do Cliente (`/portal`):**
  - Rota pública do cliente protegida por token criptografado de 32 bytes (armazenado com hash SHA-256 no banco).
  - Linha do tempo visual mobile-first em 4 fases: Requisitos, Construção, Testes e Entrega.
  - Isolamento estrito de dados: zero vazamento de custos internos, margens de lucro, prompts de IA ou notas internas confidenciais para o cliente final.
  - Atualização automática de status (polling a cada 15 segundos).
  - Modal de Aceite Formal em 2 etapas na fase de entrega, com assinatura do cliente, hash de estado imutável e registro em `audit_log`.
- **Backend Go e Armazenamento Seguro:**
  - Migração `00008_client_portal_tokens.sql` com RLS ativada e forçada.
  - Endpoints públicos `GET /v1/public/portal/{token}` e `POST /v1/public/portal/{token}/accept`.
  - Endpoints autenticados do Dono: `POST /v1/cards/{id}/portal-link`, `GET /v1/cards/{id}/portal-link` e `DELETE /v1/cards/{id}/portal-link`.
  - Suíte de testes automatizados `portal_test.go` cobrindo geração de tokens, isolamento estrito de dados financeiros e auditoria de aceite formal idempotente.
- **Integração no Painel do Dono:**
  - Widget de gestão do Portal do Cliente dentro da `CardDetailSheet` do cartão com geração de link, cópia rápida para compartilhamento, exibição de último acesso do cliente, confirmação de aceite formal e revogação imediata de acesso.

---

## [Sprint 03 - Fase 12] - Gerador de Propostas Comerciais e Contratos — 2026-10-21

### Adicionado
- **Motor de Geração de Propostas & Contratos Técnicos:**
  - Caso de uso `GenerateProposalDocument` na API Go com escopo executivo, cronograma por tipo de demanda, garantia técnica de 90 dias e condições de pagamento em BRL.
  - Integração com especificações BDD (User Stories) e ordens PIX quando existentes no cartão.
  - Endpoints autenticados `GET /v1/cards/{id}/proposal-document` e `POST /v1/cards/{id}/proposal-document`.
  - Registro de auditoria em `audit_log` e eventos imutáveis no histórico do cartão (`proposal_document_generated`).
- **Visualizador e Exportador Executivo (`ProposalModal`):**
  - Layout timbrado profissional com logo Geevheit e metadados de projeto.
  - Ação de 1 toque: **"Copiar Proposta para WhatsApp"** pré-formatada em Markdown com emojis, tópicos, link de aceite e chave PIX.
  - Ação de 1 toque: **"Imprimir / Salvar PDF"** com estilização `@media print` para exportar documento A4 de alta fidelidade sem dependências de terceiros.
  - Botão de abertura rápida integrado à `CardDetailSheet` do Dashboard.

---

## [Sprint 03 - Extra] - Identidade Visual Proprietária & Branding — 2026-10-21

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
