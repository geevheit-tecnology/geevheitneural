---
trigger: always_on
---

# GeevheitNeural — Segurança e LGPD

- Login só do dono, com 2FA (TOTP). API exige JWT válido, `aal=aal2` e `sub = OWNER_USER_ID`.
- RLS em **todas** as tabelas, aplicada também às queries da API (role `authenticated` + claims por transação).
- Nenhum segredo no frontend, no repositório ou no banco. `.env` fora do git; manter `.env.example` atualizado e sem valores reais.
- Segredos de conectores só em variáveis do servidor/cofre, referenciados por `secret_ref`.
- Conectores: `ativo=false` e escopo somente leitura por padrão.
- Todo evento relevante (decisão de portão, pausa/retomada, mudança de conector/agente, login falho) vai para `audit_log`.
- Logs e `audit_log` com IDs, não com nome, e-mail, telefone ou texto livre do cliente.
- Endpoint público: validação de esquema, limite de tamanho, honeypot, rate limit; IP não é persistido em texto puro.
- Nenhuma ação automática em produção sem portão de aprovação do dono.
- Mudança que toca auth, RLS, portões ou ExecutionGuard **exige teste** novo ou atualizado.
