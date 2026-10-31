# Correção 004 — diagnóstico e validação local (2026-10-30)

## Diagnóstico comprovado

Jarvis e Jake são personas do mesmo `JarvisChatDrawer`, com seleção de voz diferente; não existe chamada recursiva entre dois agentes. Web Speech converte voz em texto e `handleSend` chama `onExecuteCommand` → `processCommandMutation` → API Go → parser determinístico → `StartBuildUseCase`. O modelo é chamado pelos runners Groq/Gemini apenas na esteira de geração. O histórico do chat não é reenviado ao modelo.

A repetição foi reproduzida com eventos simulados atrasados: `onresult` de um reconhecimento já abortado aceitava novos textos, e `onend` antigo apagava a referência do microfone atual. As proteções anteriores contra resultados duplicados e eco foram preservadas; agora os callbacks verificam a identidade do reconhecimento ativo e o estado da síntese. Não foi reproduzido um loop acústico com aparelho real.

Falhas de geração encontradas:

- Frontend substituía qualquer erro da API por respostas locais de sucesso. O botão Construir anunciava conclusão ao receber a confirmação de início, ou mesmo após erro.
- Sem chave/configuração válida, o composition root utilizava silenciosamente o runner de simulação.
- JSON inválido do modelo era aceito como texto e saída vazia virava mensagem de sucesso.
- Plano do Arquiteto e arquivos do Dev não chegavam aos revisores. `previous_code` continha somente uma URL de PR.
- O contexto assíncrono descartava claims usados nas transações RLS.
- Escrita de arquivos, transações e publicação ignoravam erros. Ausência de arquivos produzia HTML genérico com alegações de homologação.
- QA marcava testes aprovados incondicionalmente; ficha de risco continha números fixos sem execução de ferramentas. Conclusão era marcada antes da última etapa.
- Chamadas repetidas de construir no mesmo cartão iniciavam esteiras concorrentes.

## Arquivos alterados

| Arquivo | Ajuste |
|---|---|
| `api/cmd/geevheit-api/main.go` | Ausência de provedor não seleciona simulação silenciosa. |
| `api/internal/adapters/agents/gemini_runner.go` | Parser compartilhado por Groq/Gemini rejeita JSON inválido e saída vazia. |
| `api/internal/app/pipeline_usecases.go` | Mantém contexto, encaminha artefatos, bloqueia construção concorrente por cartão na instância, exige arquivos gravados/publicação válida, propaga erros de persistência das etapas, registra falha e conclui somente após a última etapa. Remove resultados genéricos e aprovações de testes sem evidência; QA pode rejeitar a entrega. |
| `web/src/app/page.tsx` | Comandos usam resultado da API; falhas são propagadas ao chat. Construção informa início sem promover entrega localmente. |
| `web/src/shared/api-client/index.ts` | Timeout informa execução não confirmada em vez de anunciar central local. |
| `web/src/features/dashboard/components/JarvisChatDrawer.tsx` | Ignora resultados e encerramentos de reconhecimentos antigos, inclusive durante TTS. |
| `api/internal/app/sprint02_test.go` | Cenário existente passa a usar arquivos de fixture, sandbox temporária e publicação local; aguarda trabalhadores antes da limpeza. |
| `api/internal/app/correcao004_test.go` | Regressões de geração, contexto, falhas e duplicação. |
| `api/internal/adapters/agents/correcao004_test.go` | Regressões do parser de respostas. |
| `web/scripts/test-assistant.cjs` | Executa handlers reais com Web Speech simulado para ambas as personas. |
| Documentação | Este relatório, TASKS, MEMORY e CHANGELOG registram o alcance local. |

`correcao004.md` e a alteração preexistente em `web/scripts/build-css.js` foram preservados.

## Validação

- `go test -race ./...`: passou após ajustes; repositórios de banco dependem de configuração própria e esse comando não comprova integração com banco remoto.
- `node web/scripts/test-assistant.cjs`: 4 testes passaram (Jarvis/Jake, evento duplicado/atrasado, encerramento antigo e falha de API).
- `npx tsc --noEmit`: passou.
- `npm run build`: passou, exportação estática de 9 páginas.
- Testes novos verificam arquivo real em disco, plano e código no contexto dos revisores, preservação de contexto após cancelamento HTTP, saída vazia, resultado nulo, erro de provedor, escrita, publicação, rejeição de QA e início duplicado.

## Limites e pendências

A geração foi validada **localmente com runner simulado e arquivos reais em diretório temporário**. Não houve chamada real a Groq/Gemini, validação de credenciais/modelos remotos, migração, deploy, push ou alteração do banco de produção. Não havia logs de produção disponíveis nesta inspeção.

Continuam pendentes: teste de voz em celular real; geração com provedor real e persistência remota; homologação funcional do produto gerado. QA/Security continuam sendo revisões do modelo: ferramentas de testes, Lighthouse/SAST e auditoria de dependências não são executadas pela esteira atual. A correção deixa isso explícito. Bloqueio de duplicação vale para a instância em execução; reinício do processo não recupera tarefas interrompidas. O timeout existente de 3,5 s ainda pode ocorrer durante hibernação do host; a interface não declara sucesso nesse caso.
