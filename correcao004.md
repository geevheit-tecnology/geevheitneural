# MISSÃO: CORRIGIR A EXECUÇÃO REAL DOS AGENTES EXISTENTES

## CONTEXTO

O projeto é uma fábrica de software multiagente já desenvolvida em Golang.

Os agentes Jarvis e Jake já existem no projeto e possuem interação por voz.

Os problemas observados são:

1. A fábrica não gera os resultados esperados quando recebe uma solicitação.
2. Os agentes Jarvis e Jake ficam repetindo as mesmas mensagens ou respostas.
3. A interação parece ocorrer, mas o trabalho esperado não é efetivamente concluído.

O objetivo é corrigir esses comportamentos na implementação existente.

## REGRA ABSOLUTA

NÃO reconstruir o sistema.

NÃO substituir os agentes existentes.

NÃO modificar a arquitetura.

NÃO adicionar novos agentes.

NÃO trocar o provedor de IA.

NÃO reescrever o sistema de voz.

NÃO modificar a interface visual sem comprovação de que ela está relacionada ao problema.

NÃO criar funcionalidades novas.

Trabalhe exclusivamente na identificação e correção das causas dos três problemas relatados.

### ETAPA 1 — INSPECIONAR O FUNCIONAMENTO ATUAL

Antes de modificar qualquer arquivo:

1. Localize as implementações existentes do Jarvis e do Jake.
2. Identifique como cada agente recebe uma solicitação.
3. Identifique como a entrada de voz é convertida em texto, caso essa conversão seja utilizada.
4. Identifique como a solicitação chega ao modelo de IA.
5. Identifique como a resposta do modelo é processada.
6. Identifique como o agente decide concluir uma resposta ou iniciar uma ação.
7. Identifique como as tarefas são encaminhadas ao orquestrador.
8. Identifique onde os resultados das tarefas deveriam ser produzidos e armazenados.
9. Identifique como o sistema decide que uma execução foi concluída.
10. Examine os logs e os erros disponíveis.

Respeite a organização real do código. Não presuma nomes de arquivos, funções, bibliotecas ou componentes.

### ETAPA 2 — INVESTIGAR A REPETIÇÃO DE RESPOSTAS

Verifique especificamente se:

- O mesmo evento de voz está sendo processado mais de uma vez.
- Uma resposta está sendo reenviada como se fosse uma nova solicitação.
- O histórico de conversação está sendo reutilizado incorretamente.
- O agente está recebendo sempre o mesmo contexto.
- O orquestrador está retornando à mesma etapa sem avançar.
- Existe um ciclo de chamadas entre Jarvis e Jake.
- A resposta do modelo está sendo confundida com uma instrução para continuar.
- Uma tarefa concluída continua marcada como pendente.
- O tratamento de erros está provocando tentativas repetidas.
- A interface está reproduzindo mensagens antigas.
- O estado da conversa ou da execução não está sendo atualizado corretamente.

Esses itens são hipóteses de investigação, não conclusões antecipadas.

Não implemente correções para problemas que não tenham sido encontrados.

### ETAPA 3 — INVESTIGAR POR QUE A FÁBRICA NÃO GERA RESULTADOS

Descubra qual é o fluxo real de geração implementado no projeto.

Verifique se a solicitação percorre corretamente as etapas existentes:

SOLICITAÇÃO → INTERPRETAÇÃO → PLANEJAMENTO → EXECUÇÃO → PRODUÇÃO DO RESULTADO → VALIDAÇÃO → CONCLUSÃO.

Essa sequência é apenas uma referência para localizar falhas. Preserve o fluxo efetivamente implementado quando ele for diferente.

Verifique se o sistema apenas conversa com o usuário ou se realmente executa as operações necessárias para produzir os resultados esperados.

Examine:

- Chamadas reais ao provedor de IA.
- Respostas vazias ou inválidas.
- Erros de autenticação ou configuração.
- Chamadas de ferramentas não executadas.
- Resultados descartados antes de serem armazenados.
- Falhas no encaminhamento entre agentes.
- Tarefas que ficam presas em execução.
- Falhas na criação ou gravação dos arquivos de saída.
- Erros que são capturados, mas não apresentados.
- Execuções que são declaradas concluídas sem produzir o resultado esperado.

Não suponha que a geração esteja ausente. Descubra se ela já está implementada e por que não funciona.

### ETAPA 4 — CORRIGIR A CAUSA IDENTIFICADA

Implemente a menor alteração capaz de corrigir cada falha comprovada.

Preserve os agentes, suas configurações, as integrações e os contratos existentes.

Não faça refatorações gerais.

Não substitua implementações inteiras por soluções simplificadas.

Não crie respostas fixas para simular inteligência.

Não considere mensagens de voz como prova de que uma tarefa foi executada.

Uma resposta conversacional não pode ser apresentada como resultado de uma tarefa que nunca foi executada.

Se uma solicitação exigir geração de código ou de arquivos, valide se a operação foi realizada e se o resultado esperado existe.

Se a operação falhar, registre a falha e informe o estado correto da execução.

### ETAPA 5 — EVITAR REPETIÇÕES INFINITAS

Se o problema estiver relacionado a ciclos de execução, corrija o mecanismo responsável sem prejudicar as interações legítimas.

Garanta, conforme a arquitetura existente, que:

- Cada solicitação possua uma identificação rastreável.
- Os eventos duplicados possam ser identificados quando necessário.
- A mesma etapa não seja repetida indefinidamente.
- As retentativas tenham limites.
- Os erros sejam diferenciados de respostas válidas.
- As execuções possuam estados consistentes.
- O agente consiga concluir uma tarefa.
- Uma execução com falha não seja apresentada como concluída.

Não implemente controles redundantes se o sistema já possuir mecanismos equivalentes funcionando.

### ETAPA 6 — TESTAR O COMPORTAMENTO REAL

Utilize os agentes existentes e os testes disponíveis.

Valide separadamente:

*Teste A — Jarvis*
- Enviar uma solicitação.
- Verificar se ela é processada.
- Confirmar se a resposta corresponde à solicitação.
- Verificar se a mesma resposta não é repetida indevidamente.

*Teste B — Jake*
- Repetir o procedimento com o Jake.
- Confirmar que ele processa a solicitação recebida.
- Verificar se não existe repetição indevida.

*Teste C — Comunicação entre agentes*
- Verificar se a comunicação entre os agentes ocorre conforme o fluxo implementado.
- Confirmar que uma resposta anterior não provoca um ciclo infinito.

*Teste D — Geração*
- Enviar uma solicitação que corresponda a uma funcionalidade de geração já existente.
- Confirmar se o processamento realmente ocorre.
- Verificar se o artefato ou resultado esperado é produzido.
- Confirmar que o sistema informa corretamente sucesso ou falha.

*Teste E — Regressão*
- Executar os testes existentes.
- Confirmar que a correção não prejudicou os comportamentos anteriormente funcionais.

Utilize testes simulados quando apropriado, mas diferencie-os dos testes que dependem de integrações reais.

Não invente resultados de execução.

### ETAPA 7 — PRESERVAR O PROJETO

Antes de alterar arquivos, examine o estado do Git.

Não sobrescreva alterações existentes.

Não apague arquivos ou dados.

Não modifique bancos de dados ou configurações de produção sem autorização.

Não execute comandos destrutivos.

Se a correção exigir uma mudança estrutural, interrompa essa alteração e apresente a justificativa técnica.

### ETAPA 8 — RELATÓRIO FINAL

Apresente um relatório curto contendo:

1. A causa comprovada da repetição do Jarvis, se identificada.
2. A causa comprovada da repetição do Jake, se identificada.
3. A causa pela qual a geração não funciona, se identificada.
4. Os arquivos efetivamente modificados.
5. A correção aplicada em cada arquivo.
6. Os testes executados e seus resultados.
7. Se a geração real foi validada.
8. Quais problemas continuam pendentes.

Se não conseguir reproduzir uma falha, informe isso.

Se não houver evidências suficientes, não invente uma solução.

## INSTRUÇÃO FINAL

Comece exclusivamente pela inspeção do código existente e pelo diagnóstico dos três sintomas relatados.

Não modifique arquivos durante a inspeção inicial.

Depois de identificar uma causa concreta, aplique somente a correção necessária e valide seu efeito.

O objetivo não é tornar o código mais bonito nem criar uma nova fábrica.

O objetivo é fazer Jarvis, Jake e o mecanismo de geração existentes funcionarem corretamente, preservando todo o restante do sistema.
