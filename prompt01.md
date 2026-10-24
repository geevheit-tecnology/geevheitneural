# GeevheitNeural — Prompt Mestre de Execução (prompt01.md)

> **Instrução para Agentes de IA:**  
> Quando o Dono solicitar *"faça tudo que falta"* ou iniciar uma nova sessão com base neste arquivo, leia este documento na íntegra, verifique o estado atual do repositório e execute as etapas pendentes de forma autônoma, respeitando o custo zero (R$ 0,00), a arquitetura Go Clean/SOLID e a soberania do Dono nos portões 2FA.

---

## 1. Visão Geral do Sistema e Estado Atual

O **GeevheitNeural** é uma central de operação de software house autônoma e mobile-first. Um único dono decide e aprova ordens pelo celular; agentes de IA concebem, codificam, testam e publicam projetos de software sem custo de infraestrutura.

### ✅ O Que Já Está Pronto e Testado:
1. **Backend Go (`/api`):**
   - Clean Architecture e SOLID estritos (`domain` puro → `app` com casos de uso e portas → `adapters` → `cmd`).
   - `GeminiAgentRunner` (`api/internal/adapters/agents/gemini_runner.go`): Conector com Google AI Studio (Gemini Flash), rate limiter de 15 RPM e tratamento de erros.
   - `LocalSandboxManager` (`api/internal/adapters/sandbox/local_sandbox.go`): Criação isolada de diretórios de projeto com proteção contra path traversal.
   - `LocalPublisher` (`api/internal/adapters/publisher/local_publisher.go`): Inicialização de Git local, commits atômicos e endpoints de preview (`/v1/cards/{id}/preview`).
   - Casos de uso de Pipeline integrados (`StartBuildUseCase`).
   - Testes com `-race` passando 100% no backend.
2. **Frontend Next.js PWA (`/web`):**
   - Mobile-first responsivo validado em 360 px.
   - **Agenda & Prazos:** Mini calendário semanal de 7 dias com badges de vencimento (Hoje, Atrasados, Próximos 7 dias).
   - **Hub de Quadros:** Visão geral da software house com filtro por cliente e status.
   - **3 Decisões do Dia:** Fila inteligente priorizada por dinheiro em jogo.
   - **Resiliência a Falhas de Rede:** Estado reativo local que permite aprovar, recusar e enviar comandos pelo celular mesmo antes de conectar a API remota.
   - **PWA Service Worker:** `sw.js` com estratégia Network-First atualizada.
   - **Deploy de Produção:** Publicado e ativo na Vercel via [is.gd/geevheit](https://is.gd/geevheit).

---

## 2. Invariantes Invioláveis de Arquitetura

1. **Portões Exclusivos do Dono:** Agente **nunca** aprova portão (`proposta`, `prompt`, `entrega`). Somente o Dono com 2FA (`aal2`).
2. **ExecutionGuard Mandatório:** Nenhuma IA roda e nenhum conector executa se o sistema estiver pausado ou inativo.
3. **Custo Zero (R$ 0,00):** Utilizar estritamente os planos gratuitos (Free Tiers) sem custo fixo:
   - IA: Google AI Studio (15 RPM / 1.500 req/dia).
   - Banco: Supabase Postgres Free Tier (São Paulo `sa-east-1`).
   - Frontend: Vercel Free Hobby.
   - Backend: Render / Fly.io / Railway Free Tier ou túnel Cloudflare/Ngrok.
   - Cobrança: Asaas / Efí / Mercado Pago (R$ 0,00 mensalidade, taxa só no Pix recebido).
4. **Segurança:** Nenhum segredo em código ou banco. Uso de `secret_ref` e variáveis de ambiente.

---

## 3. Roteiro do Que Falta Fazer (Próxima Sessão)

Quando o Dono pedir para continuar, execute as fases nesta ordem:

### 🚀 ETAPA 1: Ligar o Banco de Dados na Nuvem (Supabase São Paulo)
* **Objetivo:** Persistir clientes, cartões, eventos e aprovações em banco real na nuvem.
* **Ações:**
  1. No Supabase Cloud (região São Paulo `sa-east-1`), criar o projeto gratuito.
  2. Aplicar as migrations do diretório `db/migrations/`:
     - `00001_schema.sql` (todas as tabelas, enums, triggers de imutabilidade).
     - `00002_security.sql` (Row Level Security em todas as tabelas).
     - `00003_fase6_pipeline.sql` (tabelas de pipeline e agentes).
     - `00004_fase11_pix_payments.sql` (ordens de pagamento Pix).
  3. Cadastrar o usuário oficial do Dono com 2FA TOTP ativado.
  4. Obter as variáveis:
     - `SUPABASE_URL`
     - `SUPABASE_ANON_KEY`
     - `SUPABASE_SERVICE_ROLE_KEY`
     - `DATABASE_URL` (para a API Go)

### 🚀 ETAPA 2: Publicar e Conectar a API Go (`/api`)
* **Objetivo:** Deixar a API Go disponível na internet com HTTPS para o frontend Next.js consumir.
* **Ações:**
  1. Hospedar a API Go no **Render** (Free Web Service) ou **Fly.io** usando o `api/Dockerfile` multi-stage distroless já existente.
  2. Configurar as variáveis de ambiente no serviço da API:
     ```env
     PORT=8080
     DATABASE_URL=postgresql://postgres:...@...supabase.co:5432/postgres
     SUPABASE_URL=https://<seu-projeto>.supabase.co
     SUPABASE_JWT_SECRET=<seu-jwt-secret>
     OWNER_USER_ID=<seu-uuid-de-dono>
     AGENT_RUNNER_TYPE=gemini
     GEMINI_API_KEY=<chave-google-ai-studio>
     ALLOWED_ORIGINS=https://web-lake-eight-gwn3vopv94.vercel.app,https://is.gd/geevheit
     ```
  3. Validar os endpoints `/healthz` e `/readyz`.

### 🚀 ETAPA 3: Conectar a Vercel com as Chaves de Produção
* **Objetivo:** Fazer o PWA na Vercel se comunicar com a API Go e com o Supabase oficial.
* **Ações:**
  1. No painel da Vercel (Projeto `web`):
     - `NEXT_PUBLIC_SUPABASE_URL` = `https://<seu-projeto>.supabase.co`
     - `NEXT_PUBLIC_SUPABASE_ANON_KEY` = `<chave-anon-supabase>`
     - `NEXT_PUBLIC_API_URL` = `https://<sua-api-go-no-render>.onrender.com`
  2. Rodar novo deploy na Vercel (`npx vercel deploy --prod --yes`).
  3. Testar login real com e-mail, senha e código TOTP de 6 dígitos no celular.

### 🚀 ETAPA 4: Conectar Gateway Pix Brasileiro (R$ 0,00 de Custo Fixo)
* **Objetivo:** Receber o sinal de projetos automaticamente e liberar o pipeline da fábrica.
* **Ações:**
  1. Configurar webhook do **Asaas** ou **Efí/Mercado Pago** apontando para `POST /v1/payments/webhook`.
  2. Testar conciliação: quando o Pix de 50% é pago pelo cliente, o cartão avança automaticamente de *Aguardando Cliente* para *Prompt para Aprovar*.

### 🚀 ETAPA 5: Conector de Notificação WhatsApp (Opcional — Grátis)
* **Objetivo:** O Dono recebe um aviso no WhatsApp no celular sempre que houver um portão pendente.
* **Ações:**
  1. Usar a Cloud API oficial do WhatsApp (Meta) com as 1.000 conversas gratuitas/mês.
  2. Mensagem padrão: *"Geevheit: Há R$ X.XXX aguardando sua decisão. [Aprovar pelo Celular](https://is.gd/geevheit)"*.

---

## 4. Checklist de Execução Rápida

Quando você reiniciar o projeto amanhã, basta dizer:
> **"Continue a execução a partir do prompt01.md"**

O agente seguirá os passos acima, validará os testes unitários (`go test -race ./...` e `npm run build`), sincronizará as migrações e configurará as integrações finais.
