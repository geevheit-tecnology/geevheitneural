# PRD — GeevheitNeural

> Central de operação de software house, mobile-first.
> Fonte: [`/sprit01`](../sprit01). Versão 0.1 — 2026-10-18. Status: **aguardando aprovação do dono**.

## 1. Visão

Um sistema web instalável (PWA) pelo qual **uma única pessoa — o dono** — opera uma software house inteira pelo celular.
O dono **só decide, aprova e dá ordens** (texto ou voz). Agentes de IA serão "funcionários" ligados no futuro;
**nesta fase todos ficam desligados** (`ativo=false`).

## 2. Persona

| Persona | Descrição | Pode |
|---|---|---|
| **Dono** (único usuário humano) | Opera tudo pelo celular, 360 px, muitas vezes com uma mão. | Tudo. É o **único** que aprova portões e retoma o sistema após Parar. |
| **Agente** (futuro, desligado) | Funcionário de IA com papel, escopo mínimo e limite de gasto. | Ler cartões do seu estágio, registrar progresso, anexar resultado, **solicitar** aprovação. **Nunca aprova.** |
| **Visitante do site** | Cliente em potencial. | Enviar pedido pelo formulário público. Aceitar proposta por link. |

## 3. Problema

Operar uma software house sozinho significa perder dinheiro em decisões paradas, não saber quais apps dão prejuízo
e não ter um lugar único e seguro para aprovar o trabalho que será, no futuro, executado por agentes.

## 4. Objetivos e não-objetivos

**Objetivos**
- O1 — Mostrar em segundos as decisões que mais valem dinheiro e permitir agir com 1 toque.
- O2 — Garantir que nada avança em portões sem aprovação explícita e imutável do dono.
- O3 — Dar visão de lucro real por app, cliente e serviço.
- O4 — Deixar o encaixe dos agentes pronto (contratos), mas **sem nenhum poder de ação** nesta fase.

**Não-objetivos (fase atual)**
- App Flutter (Fase 2 do produto — só o contrato da API fica pronto).
- Ligar qualquer agente ou conector.
- Multiusuário, times e permissões além do dono.
- Java (não introduzir).

## 5. Requisitos funcionais

| ID | Funcionalidade | Resumo | Fase de entrega |
|---|---|---|---|
| RF-01 | **Hoje** | Tela inicial com as **3 decisões do dia**, ordenadas por dinheiro em jogo, recomendação de 1 linha e botão de ação direta. | 2 |
| RF-02 | **Quadro Kanban** | Cartões por estágio, detalhes, histórico (`card_events`) e ações: aprovar, pedir mais informações, recusar. | 2 |
| RF-03 | **Meus apps** | Inventário (hospedagem, banco, repositório, domínio, status), custo, receita, resultado e alerta **"custa mais do que rende"**. | 2 |
| RF-04 | **Finanças** | Receita, custo e lucro por app, cliente e serviço; sugestão de consolidação de serviços; entrada manual e importação CSV. | 2 |
| RF-05 | **Conexões** | GitHub, Vercel, Render, Neon, Supabase, Netlify, Firebase, Docker Hub, provedores de IA. Todos desligados, chave opcional, somente leitura por padrão. | 2 |
| RF-06 | **Dinheiro parado esperando você** | Contador fixo no topo: soma dos valores dos cartões parados em portões do dono + quantidade de decisões. Tempo real. | 2 |
| RF-07 | **Barra de comando (texto/voz)** | Em todas as telas. Web Speech API pt-BR. Comandos: aprovar, recusar, status, quanto gastei/lucro. Nome ambíguo ⇒ **pergunta, nunca age no cartão errado**. Mesma API e portões. Destrutivo ⇒ confirmação. | 3 |
| RF-08 | **Ficha de risco na entrega** | Antes de aprovar entrega: testes, varredura de segredos, dependências desatualizadas, pontos de LGPD. Preenchimento manual por ora; estrutura pronta para agentes de QA/segurança. | 3 |
| RF-09 | **Botão Parar** | Pausa todos os agentes e conexões. Enquanto pausado nada executa, aviso visível, só o dono retoma. Registrado em `audit_log`. | 2 (regra de domínio na 1) |
| RF-10 | **Entrada de pedidos** | Endpoint público para o formulário do site: validação, anti-spam, rate limit; cria cartão em **Entrada**. | 4 |
| RF-11 | **Proposta e aceite do cliente** | Proposta gerada a partir do cartão, aceite do cliente por link seguro. | 4 |
| RF-12 | **Contratos dos agentes** | OpenAPI + endpoints (desligados): listar cartões do estágio, ler contexto, registrar progresso, anexar resultado, solicitar aprovação. | 5 |

### 5.1 Ciclo de vida do cartão

`Entrada → Proposta para aprovar → Aguardando cliente → Prompt para aprovar → Em execução → Entrega para aprovar → Entregue`

- Tipos: `orçamento | bug | suporte | financeiro`.
- **Portões do dono:** *Proposta para aprovar*, *Prompt para aprovar*, *Entrega para aprovar*.
  Só avançam por ação explícita do dono, registrada em `approvals` (imutável).

## 6. Requisitos não funcionais

| ID | Requisito |
|---|---|
| RNF-01 | Mobile-first, uso completo em **360 px**, tema **claro e escuro**. |
| RNF-02 | PWA instalável, **offline para leitura**. |
| RNF-03 | Hospedagem gratuita (front: Cloudflare Pages; API: host gratuito com Dockerfile; banco: Supabase). Limites e plano de migração documentados. |
| RNF-04 | Simplicidade e poucas dependências — cada dependência justificada (ver [ARCHITECTURE.md](./ARCHITECTURE.md#7-dependências)). |
| RNF-05 | Cadastrar apps reais, lançar custos e ver lucro por app em **< 5 minutos**. |
| RNF-06 | Testes automatizados da API para: regras de portões, ambiguidade de comando, pausa de emergência. CI com testes e lint. |

## 7. Segurança e conformidade (obrigatório)

- Login **somente do dono**, com **2FA** (TOTP). Cadastro público desligado.
- **RLS em todas as tabelas.** Nenhum segredo no frontend.
- Segredos de conectores só no cofre/variáveis do servidor; **nunca em texto puro no banco** (o banco guarda só a referência).
- Dados de cada cliente isolados (LGPD). Logs sem dados pessoais desnecessários.
- Todo evento relevante em `audit_log`. Aprovações **imutáveis**.
- Nenhuma ação automática em produção sem portão de aprovação do dono.
- Agente nunca aprova portões. Agente/conector com `ativo=false` ou sistema pausado ⇒ **nenhuma ação**.

## 8. Critérios de aceite (globais)

- [ ] Uso completo no celular (360 px) com tema claro e escuro.
- [ ] Testes da API (portões, ambiguidade de comando, pausa de emergência) passando no CI.
- [ ] Nenhum conector ou agente consegue agir enquanto `ativo=false` ou com o sistema pausado.
- [ ] Cadastrar apps reais, lançar custos e ver o lucro por app em menos de 5 minutos.

## 9. Fases

| Fase | Nome | Entrega |
|---|---|---|
| 1 | Fundação | Monorepo, migrations, auth, RLS, API base, CI com testes e lint. |
| 2 | Painel | Hoje, Quadro, Meus apps, Finanças, Conexões, Parar, PWA instalável. |
| 3 | Comando e risco | Barra de comando texto/voz, ficha de risco. |
| 4 | Fluxo comercial | Formulário público, proposta, aceite por link, aprovações. |
| 5 | Contratos dos agentes | OpenAPI e endpoints desligados. |

Cada fase entrega: código, testes, migrations, README de execução e resumo curto do que mudou. **Para ao fim de cada fase e aguarda aprovação.**

## 10. Perguntas em aberto

Veja [MEMORY.md › Perguntas em aberto](./MEMORY.md#perguntas-em-aberto).
