---
trigger: always_on
---

# GeevheitNeural — Regras de trabalho

## Antes de qualquer tarefa
1. Ler `docs/MEMORY.md` (estado atual, decisões, armadilhas, perguntas em aberto).
2. Localizar a tarefa em `docs/TASKS.md`. Se não existir, propor a tarefa antes de codar.
3. Consultar `docs/ARCHITECTURE.md` e `docs/SYSTEM_DESIGN.md` para a parte afetada.

## Fluxo
- Trabalhar **por fase**. Ao fim de cada fase: código, testes, migrations, README de execução e resumo curto em `docs/CHANGELOG.md`. **Parar e aguardar aprovação do dono.**
- Ambiguidade ⇒ **perguntar antes de assumir**. Registrar a pergunta em `docs/MEMORY.md`.
- Atualizar `docs/TASKS.md` (`[ ]` → `[/]` → `[x]`) e `docs/MEMORY.md` ao concluir algo relevante.
- Toda decisão de arquitetura nova vira um ADR em `docs/ARCHITECTURE.md` §6.

## Escopo
- Monorepo: `/api` (Go), `/web` (Next.js), `/db` (migrations), `/docs`.
- **Não introduzir Java.** Não construir o app Flutter (só manter o contrato da API).
- **Não inventar integrações.** API externa ⇒ somente leitura, documentada.
- Simplicidade e poucas dependências. Dependência nova ⇒ justificar em `docs/ARCHITECTURE.md` §7 **antes** de adicionar.

## Invariantes do produto (nunca violar)
- A API Go é o **único** ponto de entrada de ações (tela, voz, agente, formulário).
- Portões (*Proposta*, *Prompt*, *Entrega* para aprovar) só avançam por decisão explícita do dono, registrada em `approvals`.
- **Agente nunca aprova portão.**
- Agente/conector com `ativo=false` ou sistema pausado ⇒ nenhuma execução (sempre via `ExecutionGuard`).
- Comando com alvo ambíguo ⇒ perguntar; ação destrutiva ⇒ pedir confirmação.
- `approvals`, `card_events` e `audit_log` são imutáveis.

## Convenções
- Identificadores em inglês; UI, docs e commits em português.
- Conventional Commits.
- Dinheiro sempre em centavos (`int64`/`bigint`), nunca float.
