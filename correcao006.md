GeevheitNeural — Prompt de Conserto do Orquestrador (prompt02.md)

COMANDO DE ATIVAÇÃO Quando o Dono disser qualquer uma destas frases, execute este documento do início ao fim, sem pedir novas instruções:

"Conserte o orquestrador a partir do prompt02.md"
"Execute o prompt02.md"
"Faça o conserto"

O que o comando autoriza: inspecionar o projeto, rodar testes e comandos de diagnóstico não destrutivos, e aplicar a menor correção possível no working tree.

O que o comando NÃO autoriza: commit, push, deploy, instalar ou atualizar dependências, alterar arquivos .env, alterar banco de dados ou dados, mexer em produção. Para qualquer uma dessas ações, pare e peça autorização explícita ao Dono.

MISSÃO: FAZER O SISTEMA EXISTENTE FUNCIONAR CORRETAMENTE
1. CONTEXTO

Este projeto já existe e foi desenvolvido anteriormente.

O sistema é uma fábrica de software baseada em Golang e pode possuir integrações, componentes, agentes, interfaces e funcionalidades já implementadas.

Atualmente, preciso corrigir seu funcionamento.

O objetivo exclusivo desta tarefa é identificar e corrigir o problema que impede o funcionamento correto do sistema existente.

Não quero desenvolver outro sistema. Não quero reconstruir a aplicação. Não quero refatorar o projeto. Não quero adicionar funcionalidades. Não quero mudar a arquitetura. Não quero substituir tecnologias.

Quero preservar o trabalho existente e corrigir somente o que for necessário.

2. REGRA PRINCIPAL: NÃO MODIFICAR SEM NECESSIDADE COMPROVADA

Antes de qualquer alteração, examine o projeto para entender como ele funciona atualmente.

Não presuma que a arquitetura está errada. Não presuma que faltam componentes. Não presuma que o sistema precisa de melhorias estruturais.

O código existente é a referência para determinar o comportamento esperado.

Não modifique arquivos apenas porque considera que outra implementação seria melhor.

Não faça refatorações por preferência pessoal.

Não substitua bibliotecas, frameworks, modelos de IA ou integrações que já estejam implementados e funcionando.

Não crie novos módulos, agentes, serviços ou arquivos sem demonstrar que são indispensáveis para corrigir o problema identificado.

3. PRIMEIRA ETAPA: SOMENTE INSPEÇÃO E DIAGNÓSTICO

Inicialmente, não altere o código.

Inspecione o projeto existente e identifique:

Como a aplicação é iniciada.
Quais componentes são necessários para sua execução.
Quais configurações já existem.
Quais dependências são utilizadas.
Quais funcionalidades estão implementadas.
Quais testes estão disponíveis.
Qual é o procedimento correto para executar o sistema.
Qual erro ocorre durante a inicialização ou utilização.

Execute apenas verificações não destrutivas.

Leia os arquivos e os registros de erros disponíveis.

Se necessário, execute os testes existentes e os comandos de diagnóstico adequados ao projeto (por exemplo: cd api && go vet ./... && go test -race ./... e cd web && npm run build).

Não instale dependências, atualize versões ou altere configurações nesta etapa sem autorização.

Não exponha credenciais ou informações sensíveis nos relatórios.

Resultado desta etapa

Apresente um diagnóstico curto e objetivo contendo:

O problema encontrado.
A evidência que demonstra o problema.
O arquivo ou componente envolvido, se identificado.
A menor correção possível.
O risco da alteração proposta.

Se não conseguir reproduzir ou identificar a falha, informe isso claramente.

Não invente uma causa raiz.

4. SEGUNDA ETAPA: CORRIGIR SOMENTE O PROBLEMA IDENTIFICADO

Depois de identificar a causa, implemente apenas a correção necessária, respeitando os limites deste trabalho.

A correção deverá:

Preservar o comportamento existente.
Manter as interfaces e os contratos atuais.
Preservar as configurações existentes.
Manter as integrações já utilizadas.
Evitar alterações em componentes não relacionados ao problema.
Não remover funcionalidades existentes.
Não modificar o banco de dados sem necessidade comprovada.
Não alterar dados existentes.
Não substituir implementações completas quando uma correção localizada for suficiente.

Se o problema puder ser resolvido alterando poucas linhas, não modifique dezenas de arquivos.

Se a correção exigir uma mudança estrutural, interrompa essa parte do trabalho e apresente a justificativa técnica antes de prosseguir.

Não amplie o escopo por iniciativa própria.

5. PRESERVAÇÃO E SEGURANÇA

Antes de editar arquivos, verifique o estado do Git e identifique as alterações já existentes.

Não sobrescreva alterações realizadas anteriormente por mim ou por outros desenvolvedores.

Não execute comandos destrutivos.

Não apague arquivos, migrations, configurações ou dados.

Não recrie bancos de dados.

Não altere configurações de produção.

Não modifique arquivos de ambiente contendo credenciais.

Não faça commits, pushes, publicações ou implantações sem autorização explícita.

Se não for possível garantir que uma alteração é segura, pare e explique o risco.

6. TERCEIRA ETAPA: VALIDAR A CORREÇÃO

Depois da alteração, execute as verificações pertinentes ao problema encontrado.

Compare os resultados com o estado anterior.

Verifique:

Se a correção resolveu o erro identificado.
Se o sistema consegue executar o fluxo afetado.
Se os testes existentes continuam passando.
Se a compilação continua funcionando.
Se a alteração provocou algum problema adicional.

Não altere outras partes do sistema para fazer os testes passarem artificialmente.

Não remova testes que falham.

Não desative validações.

Não substitua integrações reais por respostas simuladas para declarar sucesso.

Se uma validação depender de credenciais ou serviços indisponíveis, registre a limitação.

Se surgir outro problema durante os testes, investigue-o antes de realizar qualquer correção adicional. Aplique o mesmo princípio de mudança mínima.

7. CRITÉRIOS DE CONCLUSÃO

O trabalho será considerado concluído quando:

O problema identificado tiver sido corrigido.
O comportamento afetado tiver sido validado.
As funcionalidades não relacionadas tiverem sido preservadas.
Os testes pertinentes tiverem sido executados.
As alterações estiverem documentadas de forma objetiva.

Não é necessário melhorar todo o projeto.

Não é necessário adicionar recursos.

Não é necessário modernizar o código.

O objetivo é restabelecer o funcionamento correto do que já existe.

8. RELATÓRIO FINAL

Ao terminar, apresente somente:

Qual era o problema.
Qual era a causa, se comprovada.
Quais arquivos foram alterados.
O que foi modificado em cada arquivo.
Quais testes foram executados e seus resultados.
Se o funcionamento foi efetivamente confirmado.
Quais pendências ainda existem.

Se nenhuma alteração for necessária ou se não houver evidências suficientes para corrigir o problema com segurança, não modifique o projeto. Explique o que precisa ser verificado.

INSTRUÇÃO FINAL

Comece inspecionando o projeto, sem modificar nenhum arquivo.

Identifique o problema concreto e proponha a menor correção possível.

Preserve todo o trabalho existente.

Não refatore, não reescreva, não amplie o escopo e não altere componentes que não estejam diretamente relacionados à falha.

Sua responsabilidade é fazer o sistema existente funcionar corretamente, não desenvolver um sistema novo.

ADENDO A: INVARIANTES DO PROJETO (não podem ser alteradas)

Leia prompt01.md. Estas regras valem durante todo o trabalho:

Custo zero (R$ 0,00): nenhuma correção pode exigir serviço pago.
Arquitetura Go Clean/SOLID existente: não mudar camadas nem contratos.
ExecutionGuard continua mandatório.
Agente NUNCA aprova portão (proposta, prompt, entrega); só o Dono com 2FA (aal2).
Nenhum segredo em código ou banco.

Se uma correção só for possível violando uma dessas regras, pare e explique.

ADENDO B: HIPÓTESES A VERIFICAR NO DIAGNÓSTICO

Estas são suspeitas comuns neste tipo de stack. NÃO são causas confirmadas. Verifique cada uma com evidência (log, teste, leitura do código). Só proponha correção para as que forem comprovadas. Descarte as demais sem alterar nada.

Banco: DATABASE_URL usa a porta 5432 direta do Supabase (pode falhar por IPv6) em vez do pooler (aws-0-sa-east-1.pooler.supabase.com)? Projeto Supabase pausado por inatividade (7 dias)?
Render Free dormindo: cold start de ~50 s sendo confundido com travamento.
CORS: ALLOWED_ORIGINS contém is.gd (encurtador, não é origem real) ou falta o domínio final da Vercel?
GeminiAgentRunner: o limiter de 15 RPM bloqueia em vez de esperar? Erros 429/503 são tratados? Há timeout por chamada? Limite diário de 1.500 req?
Chamadas externas sem context com timeout, deixando o pipeline pendurado.
Cartões presos em "em execução" depois de erro no AgentRunner, Sandbox ou Publisher.
JWT: SUPABASE_JWT_SECRET incorreto ou checagem de aal2 rejeitando tokens válidos (401/403 silenciosos).
Webhook Pix POST /v1/payments/webhook: eventos duplicados avançando o cartão duas vezes; assinatura não validada.
web/public/sw.js (Network-First) servindo cache velho ou sem timeout de rede.
Variáveis do frontend (NEXT_PUBLIC_API_URL, NEXT_PUBLIC_SUPABASE_URL, etc.) apontando para URL errada.
ADENDO C: REGRAS SOBRE GIT E ESCOPO
Não fazer commit, push nem deploy. Deixe as mudanças no working tree para o Dono revisar com git diff.
Se a causa for configuração externa (painel Supabase, Render ou Vercel, variáveis de ambiente), não tente "contornar" no código: liste exatamente o que o Dono deve mudar no painel.
Uma falha por vez: corrija, valide, e só então investigue a próxima.
ADENDO D: PERGUNTAS QUE O AGENTE DEVE RESPONDER SOZINHO

Antes de perguntar qualquer coisa ao Dono, tente descobrir por conta própria:

Qual o comando de start da API e do frontend (Dockerfile, Makefile, README, package.json).
Quais logs e arquivos de erro existem no repositório.
Qual teste ou comando reproduz o problema.
