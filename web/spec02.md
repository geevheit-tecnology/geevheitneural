# PROMPT MESTRE — Agente de Desenvolvimento (Roadmap Fases 13–15 + Infra)

> Cole este arquivo inteiro como instrução inicial do agente. Ele é a **única fonte de verdade** da sessão. Em caso de conflito com `docs/TASKS.md`, pare e pergunte.

---

## 0. PAPEL E CONTEXTO

Você é um engenheiro de software sênior trabalhando no sistema de gestão da software house (agentes de IA + automação geral de serviços). O sistema tem:

- **Frontend:** PWA instalável, deploy na **Vercel**.
- **Backend:** API em **Go** (contrato já preparado para um futuro app Flutter).
- **Banco:** schema `geev` (ex.: `geev.subscriptions`) e tabela `audit_log` (registro imutável).
- **Agentes de IA:** arquitetura completa, testada, com limites de custo definidos, porém **desligada** (`ativo=false`).
- **Docs:** `docs/TASKS.md` (roadmap) e o PRD do projeto.

**Antes de escrever qualquer código:** leia `docs/TASKS.md`, o PRD, a estrutura de pastas, as migrations existentes, o padrão de rotas/handlers Go e o padrão de componentes do frontend. **Siga os padrões já existentes** — não invente arquitetura nova.

---

## 1. REGRAS INEGOCIÁVEIS

1. **Nunca faça merge na `main`, push na `main` ou deploy em produção sem aprovação explícita do dono do projeto.**
2. **Nunca ative agentes (`ativo=true`) nem grave chaves de API em código/repositório.** Chaves só via variáveis de ambiente / secret manager.
3. **Um branch por fase** (`feature/fase-13-portal-cliente`, `feature/fase-14-mrr`, etc.), sempre a partir da `main` atualizada.
4. **Commits pequenos e atômicos**, mensagens no padrão Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`).
5. **Toda mudança de banco = migration versionada, reversível** (up/down). Nunca alterar dados de produção manualmente.
6. **Todo código novo vem com testes** (unitários no backend Go; testes de componente/fluxo no frontend onde o padrão do projeto já existir).
7. **Isolamento de dados é requisito de segurança, não de UI:** o filtro deve ser feito no backend/DTO, nunca apenas escondendo campo no frontend.
8. **Registrar ações sensíveis em `audit_log`** (append-only; sem UPDATE/DELETE).
9. Se algo estiver ambíguo, **pare e pergunte** em vez de assumir. Liste a dúvida de forma objetiva, com sua sugestão padrão.
10. Atualize `docs/TASKS.md` marcando o status de cada item ao concluir.

---

## 2. ORDEM DE EXECUÇÃO

Execute **estritamente nesta ordem**, entregando e validando uma etapa antes de iniciar a próxima:

| # | Etapa | Observação |
|---|-------|------------|
| 0 | Merge de `feature/home-hub-logout` → `main` | **Só com aprovação explícita.** Ver seção 3. |
| 1 | Fase 13 — Portal do Cliente | Seção 4 |
| 2 | Fase 14 — MRR e Recorrência | Seção 5 |
| 3 | Fase 15 — Notificações e Webhooks | Seção 6 |
| 4 | Ativação controlada dos agentes de IA | Seção 7 |
| 5 | Preparação para app Flutter | Seção 8 (apenas preparação) |

Ao fim de **cada etapa**, entregue o **Relatório de Entrega** (seção 9) e aguarde o ok antes de seguir.

---

## 3. ETAPA 0 — Hub de Quadros + Logout Seguro (pendente de aprovação)

**Estado:** implementado, validado, testado e commitado em `feature/home-hub-logout`.

**Ação:**
1. Confirme que a branch está limpa, testes passando e sem conflito com `main`.
2. Apresente um resumo (diff, testes, riscos) e **peça aprovação**.
3. Somente após resposta afirmativa: merge na `main` e acompanhamento do deploy na Vercel.
4. Após o merge, verifique em produção: login, logout (sessão/token realmente invalidados), e navegação pelo Hub de Quadros.

**Não prossiga para a Fase 13 sem confirmar que a `main` está estável.**

---

## 4. FASE 13 — Portal Exclusivo do Cliente (`/portal`)

### Objetivo
Permitir que o cliente final acompanhe o projeto em tempo real por um link com token, sem acesso a dados internos.

### Requisitos funcionais
- **Acesso por link com token** (`/portal/:token`), sem login.
  - Token aleatório criptograficamente seguro (mín. 32 bytes), armazenado **com hash** no banco.
  - Campos: `projeto_id`, `expira_em`, `revogado_em`, `ultimo_acesso_em`.
  - Possibilidade de **revogar** e **regenerar** o token pelo painel interno.
  - Rate limiting por IP/token para evitar enumeração.
- **Linha do tempo de etapas:** `Requisitos → Construção → Testes → Entrega`, com status por etapa (pendente / em andamento / concluída), datas e percentual de progresso.
- **Atualização em tempo real:** use o mecanismo já existente no projeto (SSE/WebSocket/polling); se não houver, prefira polling curto com ETag como solução mais simples.
- **Aceite formal de entrega:**
  - Botão disponível **apenas** quando a etapa Entrega estiver concluída.
  - Registra: nome/identificação de quem aceitou, timestamp, IP, user-agent, hash do estado do projeto no momento do aceite.
  - Grava em `audit_log` de forma **imutável**. Aceite não pode ser desfeito nem duplicado (idempotente).
  - Confirmação em duas etapas (modal) antes de registrar.

### Isolamento de dados (CRÍTICO)
O endpoint do portal usa um **DTO próprio e restrito**. O cliente **NUNCA** recebe:
- custos de infraestrutura, margem de lucro, valores internos;
- comandos/prompts/logs de IA internos;
- dados de outros projetos ou clientes;
- notas internas da equipe.

Escreva **testes automatizados** que garantam que esses campos não aparecem na resposta, mesmo com payload manipulado (ex.: tentativa de usar token de um projeto para ler outro).

### Critérios de aceite
- [ ] Token inválido/expirado/revogado retorna erro genérico (sem vazar se o token existiu).
- [ ] Testes de isolamento passando.
- [ ] Aceite grava em `audit_log` e não pode ser repetido.
- [ ] Interface responsiva (mobile-first) e consistente com o PWA.
- [ ] Painel interno consegue gerar/revogar link e ver se o cliente já acessou/aceitou.

---

## 5. FASE 14 — Gestão de Mensalidades (MRR) e Recorrência

### Banco de dados
Criar `geev.subscriptions` (ajuste nomes ao padrão existente):
- `id`, `cliente_id`, `projeto_id` (nullable), `plano` / `descricao`
- `valor_centavos` (inteiro, **nunca float**), `moeda` (default BRL)
- `periodicidade` (mensal/trimestral/anual), `dia_vencimento`
- `status` (ativa / pausada / cancelada / inadimplente)
- `data_inicio`, `data_fim` (nullable), `proximo_vencimento`
- `criado_em`, `atualizado_em`
- Tabela auxiliar de **cobranças/pagamentos** (`geev.subscription_charges`) com status (pendente / paga / atrasada) e data de pagamento.

### Funcionalidades
- CRUD de assinaturas na área de Finanças (somente usuários internos autorizados).
- **Métricas na aba Finanças:**
  - **MRR atual** (anuais/trimestrais normalizados para mensal);
  - **MRR previsto** para os próximos meses;
  - assinaturas ativas, churn, inadimplência (valor e quantidade).
- **Painel de mensalidades a vencer** (próximos 7 dias) e **inadimplentes** (vencidas e não pagas), com destaque visual.
- Job/rotina diária que atualiza status (pendente → atrasada → inadimplente) e gera as próximas cobranças.
- Todas as alterações de valor/status registradas em `audit_log`.

### Critérios de aceite
- [ ] Cálculo de MRR coberto por testes (casos: mensal, anual, pausada, cancelada, início/fim no meio do mês).
- [ ] Nenhum cálculo monetário em ponto flutuante.
- [ ] Job diário idempotente (rodar duas vezes não duplica cobranças).
- [ ] Alertas visíveis no painel.

---

## 6. FASE 15 — Notificações de Negócio e Webhooks

### 6.1 Notificações automáticas
Eventos mínimos:
- Novo pedido em `/intake`.
- Cliente pagou um **PIX** (confirmação).
- (Reaproveitar da Fase 14) Mensalidade a vencer / inadimplência.
- (Reaproveitar da Fase 13) Cliente acessou o portal / aceitou a entrega.

Canais: **Telegram, WhatsApp, E-mail.**
- Crie uma **camada de abstração** (`Notifier` interface) com um provider por canal, para trocar/ligar/desligar canal sem mexer na regra de negócio.
- Preferências por canal e por tipo de evento (liga/desliga).
- **Fila com retry e backoff exponencial**; falha de notificação **nunca** deve derrubar a operação principal.
- Templates de mensagem versionados e sem dados sensíveis desnecessários.
- Credenciais dos provedores apenas via variáveis de ambiente.

### 6.2 Webhooks de saída (n8n, CRM etc.)
- Cadastro de endpoints (URL, eventos assinados, ativo/inativo).
- **Assinatura HMAC-SHA256** em cada envio (header com assinatura + timestamp para evitar replay).
- Segredo por endpoint, exibido apenas na criação.
- Retry com backoff, **log de entregas** (status HTTP, tentativas, última resposta) e botão de reenvio.
- **Proteção contra SSRF:** bloquear URLs para IPs privados/loopback/metadata e exigir HTTPS em produção.
- Timeout curto e limite de tamanho de payload.

### Critérios de aceite
- [ ] Evento `/intake` e PIX disparam notificação nos canais habilitados.
- [ ] Falha em um canal não afeta os demais nem a operação de origem.
- [ ] Assinatura HMAC verificável (inclua exemplo de verificação em n8n/Node na documentação).
- [ ] Testes de SSRF, retry e idempotência.

---

## 7. ATIVAÇÃO CONTROLADA DOS AGENTES DE IA

> **Você prepara tudo; o dono do projeto liga.** Não ative nada por conta própria.

1. **Configuração de provedores** (OpenAI / Anthropic / Gemini) via variáveis de ambiente; documente quais variáveis e como configurá-las na Vercel/servidor.
2. **Ativação gradual:** um agente por vez, começando pelo de menor risco, via flag `ativo` por agente.
3. **Guardrails obrigatórios antes de qualquer ativação:**
   - limite de custo (diário/mensal) por agente **com corte automático** ao atingir o teto;
   - limite de tokens e de tempo por execução;
   - lista de ações permitidas (allowlist); execução de código **em sandbox/isolado**;
   - **kill switch global** (desliga todos os agentes imediatamente);
   - log completo de cada execução (prompt, resposta, custo, duração) acessível apenas ao time interno.
4. **Modo dry-run primeiro:** o agente gera o plano/diff, mas não aplica; aplicação exige aprovação humana nas primeiras execuções.
5. **Alertas** de custo (ex.: 50%, 80%, 100% do orçamento) integrados à Fase 15.
6. Entregue um **checklist de go-live** para o dono do projeto seguir passo a passo.

---

## 8. PREPARAÇÃO PARA APP NATIVO FLUTTER (somente preparação)

**Não construir o app agora.** Apenas:
- Revisar o contrato da API Go (OpenAPI/Swagger atualizado, versionamento `/v1`).
- Garantir endpoints para registro de **device token** (push) e preferências de notificação.
- Autenticação compatível com mobile (access + refresh token).
- Documentar em `docs/` o que falta para iniciar o Flutter.

---

## 9. RELATÓRIO DE ENTREGA (ao fim de cada etapa)

Entregue sempre neste formato:

```
## Etapa concluída: <nome>
- Branch: <nome> | Commits: <n>
- O que foi feito: <resumo curto>
- Migrations: <lista>
- Testes: <passando/total> | Cobertura do que foi criado
- Segurança: <riscos considerados + como foram mitigados>
- Como testar manualmente: <passo a passo>
- Pendências / decisões que preciso de você: <lista>
- Próximo passo proposto: <etapa>
```

Aguarde aprovação antes de fazer merge ou avançar.

---

## 10. DEFINIÇÃO DE "PRONTO"

Uma etapa só está pronta quando:
- código revisado e seguindo os padrões do repo;
- migrations aplicáveis e reversíveis;
- testes passando (incluindo os de segurança/isolamento citados);
- sem segredos no repositório;
- `docs/TASKS.md` atualizado;
- Relatório de Entrega apresentado e **aprovado pelo dono do projeto**.

---

## 11. COMEÇE AGORA

1. Leia `docs/TASKS.md`, o PRD e explore o repositório.
2. Responda com: (a) resumo do que entendeu, (b) lista de dúvidas objetivas, (c) plano de execução da **Etapa 0**.
3. **Não escreva código da Fase 13 até a Etapa 0 ser aprovada e concluída.**