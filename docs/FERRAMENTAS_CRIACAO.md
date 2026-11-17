# Ferramentas de criação — APIs gratuitas e Figma manual

Data: 2026-11-16. Escolha do dono: **APIs gratuitas + Figma como apoio manual**.

## Entrega local

Em **Conexões → Ferramentas de criação**, o dono consulta a presença das chaves no servidor e seleciona Gemini ou Groq para as etapas da fábrica. Uma ferramenta de geração fica selecionada por vez. A seleção é persistida na tabela `geev.connectors` existente, com auditoria na mesma transação; não foi criada migration e nenhum banco foi alterado durante o desenvolvimento.

A API exige sessão do dono com AAL2. A ativação exige chave configurada, confirmação de conta/modelo gratuitos e sistema fora da pausa. Desligar continua permitido durante pausa. O runner verifica a preferência e o ExecutionGuard antes de cada etapa; erro na leitura interrompe a execução. Antes de iniciar um build, a disponibilidade é conferida para evitar alterar o cartão quando a geração está desligada. Alterações não cancelam uma chamada de IA já em andamento.

Antes da primeira preferência salva pelo painel, preserva-se a seleção anterior por `AGENT_RUNNER_TYPE` e pelas chaves existentes (Groq tem precedência quando o tipo está vazio, como antes). Uma segunda chave presente não ativa automaticamente outro provedor. Depois da primeira escolha, a preferência persistida passa a controlar a fábrica, inclusive após reiniciar a API.

Nenhuma chave é enviada à interface ou gravada no banco. Reutilizam-se `GEMINI_API_KEY` / `GEMINI_MODEL` e `GROQ_API_KEY` / `GROQ_MODEL` já existentes. A tela informa presença da chave, sem afirmar que conexão, saldo ou faturamento foram verificados. Não se implementou medição de saldo nem seleção dinâmica de modelos.

O Figma Make tem campo de briefing, botão para copiar e link de abertura. O dono cola o texto, gera e revisa o protótipo manualmente. Não há login, geração, importação de código ou leitura automática de arquivo Figma pela central. O modo demonstração não consulta nem ativa as APIs reais pelo novo painel.

## Pesquisa em fontes oficiais

| Opção | Uso e condição verificados | Situação nesta entrega |
|---|---|---|
| [Gemini API](https://aistudio.google.com/docs/pricing) | Tokens gratuitos para determinados modelos; condições e cotas dependem da conta/modelo. A documentação distingue explicitamente Free e Paid. | Executor existente reaproveitado e selecionável pelo painel. |
| [Groq](https://console.groq.com/docs/rate-limits) | Plano Free com cotas de requisições/tokens. Para `openai/gpt-oss-120b`, a tabela consultada indica 30 RPM, 1.000 RPD, 8.000 TPM e 200.000 TPD; a própria documentação recomenda conferir os limites reais da organização. | Executor existente reaproveitado e selecionável pelo painel. |
| [Figma Make / Starter](https://www.figma.com/pricing/) | Starter anuncia 150 créditos de IA por dia, até 500 por mês. Os créditos são compartilhados por recursos de IA e o gasto de uma geração varia. | Apoio manual para protótipos. Não precisa de chave na central. |
| [OpenRouter Free](https://openrouter.ai/collections/free-models) | Modelos `:free` e `openrouter/free`; a documentação informa 20 requisições/minuto e 50/dia sem compra de créditos para variantes gratuitas. | Candidato futuro; nenhuma integração ou chave adicionada nesta entrega. |
| [Hugging Face Inference Providers](https://huggingface.co/docs/inference-providers/pricing) | Na página consultada, créditos mensais incluídos são para PRO/Team/Enterprise; contas Free são orientadas a comprar créditos ou fazer upgrade. | Não recomendado aqui como base de geração API gratuita garantida. |

Limites são um retrato da consulta, não promessa de disponibilidade futura. Uma landing page consome contexto, saída e revisões nas várias etapas; não é possível garantir que uma cota gratuita conclua qualquer pedido sem uma geração real. O plano/modelo precisa ser conferido no provedor. A confirmação da tela é declaração do dono, **não uma trava de faturamento do provedor**. A central não habilita cobrança nem compra créditos. Erros dos runners são propagados; este seletor não muda automaticamente de provedor. Retentativas e fallback de modelos já existentes no Gemini foram preservados.

## Arquivos e contratos

- `api/internal/app/factory_tools.go`: seleção, ativação, disponibilidade e encaminhamento para os runners existentes.
- `api/internal/adapters/postgres/factory_tools.go`: persistência usando `ListConnectors` / `UpsertConnector` já gerados pelo sqlc; ID derivado do dono. RLS já existente mantém o isolamento.
- `api/internal/app/ports.go` e `api/internal/adapters/postgres/unit_of_work.go`: porta e injeção do repositório transacional.
- `api/cmd/geevheit-api/main.go`: registra executores configurados, injeta a seleção na fábrica e protege as duas rotas com o middleware do dono.
- `api/internal/app/pipeline_usecases.go`: verificação de disponibilidade anterior à alteração do cartão, preservando a correção 006.
- `GET /v1/factory/tools`: lista provedores, `configured` e `active`.
- `PUT /v1/factory/tools/{gemini|groq}`: recebe `active` e `free_tier_confirmed`; ativação exige confirmação verdadeira. Nunca recebe chave ou URL arbitrária.
- `api/openapi.yaml`: contratos das rotas.
- `web/src/features/dashboard/components/FactoryToolsPanel.tsx`, `ConexoesTab.tsx`, `web/src/app/page.tsx`, `web/src/shared/api-client/endpoints.ts`: interface e integração. `compiled.css` regenerado pelo script existente para os estilos novos.
- Testes: `api/internal/app/factory_tools_test.go` e `web/scripts/test-factory-tools.mjs`.

## Validação e limites

- `go vet ./... && go test -race ./... && go build ./...`: passou. Testes de seleção com fakes verificam autorização AAL2, confirmação gratuita, chave ausente, seleção conservada em nova instância, auditoria, desativação, pausa, indisponibilidade de estado e propagação de erro sem troca de provedor. Não equivalem a integração do novo repositório com PostgreSQL real.
- `node scripts/test-factory-tools.mjs`: 4 testes passaram (ativação condicionada, erro sem conexão fictícia, demonstração bloqueada e briefing manual/erro do clipboard).
- `node scripts/test-assistant.cjs`: 12 testes existentes passaram.
- `tsc --noEmit`: passou. `npm run build:css`: passou.
- `npm run build`: compilação, TypeScript e geração das 9 páginas concluíram, mas o comando saiu com código 1 pelo `SecurityFatalError` da versão instalada do Next.js. Não foi desativada a verificação e não foram atualizadas dependências. Após pequeno ajuste de filtragem de conectores, TypeScript e CSS foram revalidados; não se repetiu o build completo bloqueado.
- Sem teste visual em navegador/viewport físico nesta entrega. Os testes da interface usam os controles/handlers reais com doubles locais de hooks e API.
- Nenhuma geração remota, credencial, plano ou saldo foi validado; nenhum serviço foi ativado. Sem `.env`, migration aplicada, commit, push ou deploy.

## Próxima homologação

### Atualização de preparação para produção — 2026-11-16

- Dono autorizou deixar a entrega funcionando em produção, incluindo as correções e publicação necessárias.
- Next.js e eslint-config-next atualizados de 16.4.2 para 16.4.6; correções transitivas compatíveis aplicadas. `npm audit fix`: **0 vulnerabilidades**. Build completo passou, sem desativar verificação de segurança.
- PostgreSQL configurado validou ativação/troca/desligamento, auditoria JSONB no protocolo simples e isolamento AAL1/outro dono. Transação sintética terminou em **ROLLBACK confirmado**, sem dados persistidos. O primeiro teste tentou a configuração local padrão por causa do diretório do pacote; repetição com URL explicitamente passada em memória validou a conexão correta.
- `TestFactoryToolsLiveLandingPage`: sete etapas com Gemini real, sandbox e Git reais; HTML de 4.226 bytes e portão `entrega_aprovar` preservado (26,84 s). Repositórios em memória; não comprova comando autenticado na API de produção.
- Navegador Chrome com perfil temporário conferiu login sem sessão e painel de ferramentas em demonstração. Corrigido transbordamento do cabeçalho com quebra de linha dos controles: largura passou a respeitar 360 e 390 px; também conferido em 1366 px, com captura em tema escuro. Botões reais permaneceram bloqueados na demonstração. Não foi utilizado login/2FA real do dono.
- ESLint do painel, componentes de Conexões/cabeçalho e novo teste: zero erros, um aviso legado de imagem. Lint global encontrou erros em arquivos anteriores (scripts CommonJS, hooks e tipos de comandos); não foram desativadas regras nem alterados esses fluxos para mascará-los. O teste novo foi migrado para ESM (`test-factory-tools.mjs`).
- Publicação confirmada: implementação `7dcc1ea` no GitHub, Render com rota tools protegida (401) e healthz/readyz 200; preflight oficial 204. Vercel `dpl_DPgjQswVmuNjyk3ey1S51gF4xmZX` Ready, alias `https://geevheit.vercel.app`, Next.js 16.4.6 no build remoto. O resultado anterior `SecurityFatalError` foi resolvido. Sessão AAL2 real do dono não foi utilizada.

Após revisão e autorização de publicação, usar conta comprovadamente Free e modelo com cota gratuita, selecionar a API no painel e pedir uma landing page pequena pelo comando `construir`. Conferir arquivos, conteúdo, prévia e falhas reais de cota. Login/2FA do dono, persistência remota e comportamento visual móvel precisam ser verificados nesse ambiente. A publicação pública do site gerado permanece uma etapa separada.
