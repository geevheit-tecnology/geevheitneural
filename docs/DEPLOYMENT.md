# Guia de Deploy em Produção — GeevheitNeural

Este documento contém o checklist passo a passo para o deploy da plataforma GeevheitNeural em produção, seguindo as decisões de arquitetura e segurança (ADR-001 a ADR-005).

---

## 📋 Pré-requisitos para o Dia do Deploy

- [ ] Acesso à conta do [Supabase](https://supabase.com).
- [ ] Acesso à conta do [Render](https://render.com) (plano gratuito para a API Go).
- [ ] Acesso à conta da [Vercel](https://vercel.com) ou [Cloudflare Pages](https://pages.cloudflare.com) para o frontend estático.
- [ ] Repositório no GitHub sincronizado e com testes verdes (`make check`).

---

## Passo 1 — Banco de Dados e Auth: Supabase Nuvem (São Paulo)

1. **Criar Novo Projeto no Supabase:**
   - **Nome:** `GeevheitNeural-Prod`
   - **Região:** `São Paulo (sa-east-1)` *(Decisão 4 do Dono — baixa latência nacional)*.
   - Defina uma senha forte para o banco de dados e guarde-a com segurança.

2. **Configurar Autenticação e 2FA (MFA):**
   - Acesse **Authentication** $\rightarrow$ **Providers** e mantenha apenas Email ativado.
   - Acesse **Authentication** $\rightarrow$ **MFA** e certifique-se de que TOTP está habilitado.
   - Crie o usuário inicial do Dono (`dono@geevheit.com`).
   - Copie o **User UID** gerado (será o `OWNER_USER_ID`).

3. **Coletar Credenciais do Supabase:**
   - **Project URL:** `https://[PROJECT-REF].supabase.co`
   - **Anon Public Key:** Em *Project Settings* $\rightarrow$ *API*.
   - **JWT Secret:** Em *Project Settings* $\rightarrow$ *API* $\rightarrow$ *JWT Settings* $\rightarrow$ *JWT Secret*.
   - **Connection String (URI):** Em *Project Settings* $\rightarrow$ *Database* $\rightarrow$ *Connection String* (selecione modo **URI** ou **Transaction Pooler** com `sslmode=require`).

---

## Passo 2 — Executar Migrações do Banco de Dados

Todas as migrações SQL puras com Row Level Security (RLS) forçada e triggers imutáveis estão versionadas em `db/migrations/`.

No seu terminal local, execute:

```bash
# Exemplo com a Connection String do Supabase Nuvem
export PROD_DB_URL="postgresql://postgres.[PROJECT-REF]:[SUA-SENHA]@aws-0-sa-east-1.pooler.supabase.com:6543/postgres?sslmode=require"

# Aplica todas as migrations (00001 a 00006)
$(go env GOPATH)/bin/goose -dir db/migrations postgres "$PROD_DB_URL" up

# Confirma o status
$(go env GOPATH)/bin/goose -dir db/migrations postgres "$PROD_DB_URL" status
```

---

## Passo 3 — Deploy da API Go no Render

O repositório já possui um `render.yaml` e um `api/Dockerfile` multi-stage com imagem distroless não-root.

1. **No Render:**
   - Clique em **New +** $\rightarrow$ **Blueprint** (ou **Web Service** apontando para o repositório).
   - O Render lerá o `render.yaml` na raiz do projeto.
2. **Definir Variáveis de Ambiente no Painel do Render:**
   - `PORT`: `8085`
   - `LOG_LEVEL`: `info`
   - `DATABASE_URL`: `[Connection String do Supabase Nuvem com sslmode=require]`
   - `OWNER_USER_ID`: `[UID do Dono obtido no Passo 1]`
   - `SUPABASE_URL`: `https://[PROJECT-REF].supabase.co`
   - `SUPABASE_JWT_SECRET`: `[JWT Secret com mais de 32 bytes do Supabase]`
   - `CORS_ALLOWED_ORIGINS`: `https://[SEU-DOMINIO-VERCEL-OU-CLOUDFLARE]`
   - `BASE_URL`: `https://[SEU-DOMINIO-VERCEL-OU-CLOUDFLARE]`
3. **Validar Execução:**
   - Acesse `https://geevheit-api.onrender.com/healthz` $\rightarrow$ deve retornar `{"status":"healthy"}`.
   - Acesse `https://geevheit-api.onrender.com/readyz` $\rightarrow$ deve retornar `{"status":"ready"}` confirmando conexão com o Postgres.

---

## Passo 4 — Deploy do Frontend na Vercel ou Cloudflare Pages

O frontend Next.js 14 está configurado com `output: 'export'` gerando HTML/JS/CSS 100% estático em `web/out/`.

1. **Na Vercel:**
   - Importe o repositório GitHub.
   - **Root Directory:** selecione `web`.
   - **Framework Preset:** Next.js.
   - **Build Command:** `npm run build`
   - **Output Directory:** `out`
2. **Definir Variáveis de Ambiente no Painel da Vercel:**
   - `NEXT_PUBLIC_SUPABASE_URL`: `https://[PROJECT-REF].supabase.co`
   - `NEXT_PUBLIC_SUPABASE_ANON_KEY`: `[Anon Key do Supabase]`
   - `NEXT_PUBLIC_API_URL`: `https://geevheit-api.onrender.com`
3. **Deploy:**
   - Clique em **Deploy**. A Vercel gerará o build e distribuirá na CDN global.

---

## Passo 5 — Checklist de Verificação Pós-Deploy

1. **Acesso do Dono:**
   - Abrir o domínio da Vercel pelo celular (viewport 360px).
   - Realizar login com email, senha e validação do código 2FA (TOTP).
2. **Pausa de Emergência:**
   - Testar o botão vermelho "Parar Sistema" no topo $\rightarrow$ verificar banner e bloqueio via `ExecutionGuard`. Retomar o sistema.
3. **Comando e Fábrica:**
   - Testar a barra de comandos ("status", "lucro").
   - Verificar na aba Finanças o "Placar da Fábrica de Software" e a "Biblioteca de Playbooks".
4. **Fluxos Públicos do Cliente:**
   - Enviar um lead de teste em `/intake`.
   - Conferir se o cartão aparece na aba "Hoje" para o dono gerar proposta comercial.
