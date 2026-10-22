# PROMPT: Tela Inicial (Hub de Quadros) + Botão de Sair

## 0. Contexto

Este projeto é um **orquestrador de agentes para desenvolvimento de software** com interface **estilo Miro**: um **canvas infinito com quadros (boards)**, onde ficam **cartões/nós** (tarefas, agentes, etapas de fluxo) que recebem **etiquetas de produção** (ex.: Backlog, Em produção, Em revisão, Bloqueado, Entregue).

Ele será usado para **gerenciar toda uma empresa de tecnologia e desenvolvimento**: vários projetos, equipes de agentes, fluxos de produção, execuções e entregas.

Faltam dois itens essenciais:

1. **Tela inicial (Home / Hub de Quadros)**: ponto de entrada após o login, para achar, abrir, criar e acompanhar quadros e a produção da empresa.
2. **Botão de sair (Logout)**: encerramento seguro da sessão, **sem perder nada do que está no canvas**.

**REGRA DE OURO: NÃO QUEBRAR NADA DO QUE JÁ FUNCIONA.**
Esta tarefa é **aditiva**. Não refatore, não renomeie, não mova e não apague código existente, a não ser o estritamente necessário para integrar os dois itens. **Atenção especial ao canvas, ao modelo de dados dos quadros, cartões e etiquetas, e à execução dos agentes: são o coração do sistema e não podem ser alterados de forma incompatível.**

---

## 1. Fase 0: Descoberta (SOMENTE LEITURA, não altere nada ainda)

Antes de escrever código, analise o projeto e produza `docs/DESCOBERTA_HOME_LOGOUT.md` com:

- **Stack** (linguagem, framework, backend, banco, gerenciador de pacotes).
- **Tecnologia do canvas**: biblioteca usada (ex.: React Flow, tldraw, Konva, Excalidraw, fabric.js, canvas/SVG próprio) e como o canvas é montado/desmontado.
- **Modelo de dados**: como são armazenados **quadros, cartões/nós, conexões, etiquetas de produção, agentes, execuções**. Onde ficam (banco, arquivos, localStorage) e quais campos existem (status, responsável, datas).
- **Etiquetas de produção**: lista atual, cores, onde são definidas, se são fixas ou configuráveis, e como um cartão guarda sua etiqueta.
- **Salvamento do canvas**: é automático (autosave) ou manual? Com que frequência? Existe estado "não salvo"? Existe colaboração em tempo real (WebSocket/SSE/CRDT)?
- **Roteamento**: rotas existentes, rota raiz (`/`) atual e para onde leva (provavelmente direto para um quadro).
- **Layout/navegação**: sidebar, barra superior, menus existentes.
- **Autenticação**: existe login? Como a sessão é guardada? Existe guard de rotas? Existe endpoint de logout?
- **Sistema de UI**: componentes, CSS/Tailwind, tema, ícones.
- **Testes**: existentes e comando exato para rodar.
- **Riscos**: pontos onde a mudança pode quebrar algo.

**Ao final, rode testes e build existentes e registre o resultado como linha de base.**

---

## 2. Regras de segurança (obrigatórias)

1. Trabalhe na branch nova **`feature/home-hub-logout`**. Nunca direto na `main`/`master`.
2. **Commits pequenos e atômicos**, com mensagens claras.
3. **Não altere** esquemas de banco de forma destrutiva. Migrações apenas **aditivas e reversíveis**.
4. **Não altere** contratos de API existentes. Se precisar de dados novos, **crie endpoints novos, somente leitura quando possível**.
5. **Não altere** o formato em que quadros, cartões e etiquetas são salvos. Quadros já existentes devem abrir exatamente como antes.
6. **Não instale dependências novas** sem necessidade real e sem justificar no relatório.
7. **Siga as convenções do projeto** e **reutilize** componentes, hooks e estilos existentes.
8. **Nenhum dado falso escondido.** Dados que não existirem devem aparecer como estado vazio explícito, nunca como números inventados.
9. **Nenhum segredo** em código ou logs.
10. Em caso de ambiguidade com risco de quebra, **pare e registre a dúvida** em vez de adivinhar.

---

## 3. Entregável A: Tela Inicial (Hub de Quadros)

### 3.1 Objetivo
Ser o "painel de comando" da empresa: em poucos segundos o usuário vê **onde está a produção**, **retoma o trabalho de onde parou** e **abre ou cria quadros**.

### 3.2 Rota e comportamento
- Rota principal `/` (ou equivalente), **protegida por autenticação**. Após o login, o usuário cai aqui.
- Se hoje `/` abre direto um quadro, **preserve esse acesso** (via "Continuar de onde parei" e pelos cartões de quadro) e documente a mudança.
- Item **"Início"** como **primeiro item** do menu de navegação, destacado quando ativo.
- **O canvas não deve ser carregado na Home.** Nenhuma biblioteca pesada do canvas deve entrar no carregamento inicial (usar carregamento sob demanda), para a Home abrir rápido.

### 3.3 Seções

**a) Cabeçalho**
- Saudação ("Bom dia, {nome}"), data e busca global (quadros, cartões, agentes). A busca por cartões é opcional e só deve ser feita se o backend já permitir.
- Botão primário **"+ Novo quadro"**.

**b) Continuar de onde parei**
- Últimos quadros abertos pelo usuário, com **miniatura (preview)** do canvas, nome, projeto, "editado há X" e quem editou por último.
- Se não houver como gerar miniatura, usar um cartão com ícone e metadados. **Não criar processo pesado de geração de imagem sem aprovação.**

**c) Painel de produção (etiquetas)**
- **Contadores por etiqueta de produção** somando todos os quadros que o usuário pode ver (ex.: Backlog 12 · Em produção 8 · Em revisão 3 · Bloqueado 2 · Entregue 40), usando **as mesmas cores e nomes das etiquetas já existentes**.
- Clicar numa etiqueta mostra a **lista de cartões** daquela etiqueta (título, quadro de origem, responsável/agente, tempo na etapa) e permite **abrir o quadro já posicionado naquele cartão**, se o canvas suportar isso. Caso contrário, abre o quadro.
- Destaque visual para **itens Bloqueados** e **cartões parados há muito tempo** na mesma etiqueta.
- Se as etiquetas forem configuráveis, o painel deve **refletir automaticamente** as configuradas, sem lista fixa no código.

**d) Todos os quadros**
- Grade/lista de quadros com: nome, projeto, miniatura, última edição, **resumo de etiquetas** (mini-barra de progresso por etiqueta) e participantes.
- **Filtros**: por projeto, por etiqueta, por responsável, por "meus quadros / compartilhados comigo / todos".
- **Ordenação**: recentes, nome, mais ativos.
- **Favoritos** (estrelinha), se houver onde persistir sem alterar estrutura existente (preferir tabela/coluna nova e aditiva).
- Alternância entre visualização em **grade** e **lista**.

**e) Atividade recente**
- Últimas execuções de agentes e mudanças de etiqueta (ex.: "Agente X moveu *Tela de login* para *Em revisão*"), com horário relativo e link para o quadro/cartão.

**f) Alertas e pendências**
- Falhas de execução, agentes travados, itens bloqueados, tarefas aguardando aprovação humana, limites de custo próximos. Ordenados por severidade.

**g) Indicadores gerais**
- Quadros ativos, agentes (ativos/ociosos/com erro), execuções do dia e custo/tokens (se houver dados).

**h) Saúde do sistema**
- Operacional / degradado / fora do ar para API, fila/worker, banco e integrações.

**i) Ações rápidas**
- **Novo quadro**, **Novo quadro a partir de modelo** (somente se já existir sistema de templates), **Criar/configurar agente**, **Ver logs**, **Configurações**. Só exiba ações cujas telas **existem de fato**.

### 3.4 Criação de quadro
- O botão "Novo quadro" deve usar **exatamente o mesmo fluxo/serviço que o sistema já usa** para criar quadros. Se não existir fluxo de criação separado, criar um modal simples (nome, projeto) que chame a lógica já existente.
- Após criar, abrir o quadro novo.

### 3.5 Requisitos de qualidade
- **Estados completos** por seção: carregando (skeleton), vazio (com orientação, ex.: "Crie seu primeiro quadro"), erro (com "Tentar novamente"). **Erro em uma seção não derruba a página.**
- **Responsivo** (desktop, tablet, mobile) e **acessível** (foco visível, teclado, `aria-label`, contraste, **não depender só de cor para diferenciar etiquetas**: incluir texto/ícone).
- **Tema** claro/escuro se o projeto já suportar.
- **Textos em pt-BR** (ou idioma já usado).
- **Desempenho**: seções carregam **em paralelo e de forma independente**; listas longas com paginação ou carregamento incremental; miniaturas com *lazy loading*.
- **Atualização**: botão "Atualizar" e atualização leve automática (30–60 s) em atividade, alertas e contadores, sem piscar a tela.
- **Permissões**: respeitar papéis (admin, gestor, dev etc.). O usuário só vê quadros e cartões que tem permissão para ver.

### 3.6 Camada de dados
- Serviço/hook dedicado (ex.: `homeService` / `useHome`) concentrando as consultas da Home.
- Agregações de etiquetas devem ser feitas **no backend** (ex.: `GET /api/home/summary`, `GET /api/home/labels`), com consulta eficiente (agrupamento no banco), e **nunca carregando todos os quadros completos no navegador** para contar cartões.

---

## 4. Entregável B: Botão de Sair (Logout)

### 4.1 Posição e visual
- **Menu do usuário** (avatar/nome → dropdown) na barra superior, com Perfil (se existir), Configurações (se existir) e **Sair**.
- O menu do usuário e o **Sair** devem estar acessíveis **também dentro da tela do quadro/canvas** (na barra superior existente), não só na Home.
- Também disponível no menu mobile.
- Ícone de saída + texto "Sair", seguindo o estilo existente.

### 4.2 Comportamento (completo e seguro)

**Prioridade máxima: não perder trabalho no canvas.**
1. Ao clicar em **Sair**:
   - Se estiver em um quadro com **alterações não salvas** ou salvamento em andamento, **forçar o salvamento (flush do autosave) e aguardar a conclusão**. Se falhar, **avisar o usuário** ("Há alterações que não foram salvas. Sair mesmo assim?") com opções **Cancelar** e **Sair sem salvar**.
   - Se não houver pendências, abrir confirmação curta: "Deseja realmente sair?" (Cancelar / Sair).
2. Ao confirmar:
   - Chamar o **endpoint de logout** no backend para invalidar a sessão/refresh token. Se não existir, **criar** (`POST /api/auth/logout`) compatível com a autenticação atual.
   - **Encerrar conexões em tempo real** (WebSocket/SSE/colaboração) e **remover a presença** do usuário nos quadros abertos, sem corromper o estado do quadro para os demais participantes.
   - **Cancelar** requisições em andamento e **parar timers** de autosave e de atualização.
   - **Limpar o estado local de sessão**: tokens, cookies, dados do usuário, stores globais, caches de quadros e de consultas, e **rascunhos locais do canvas** que pertençam à sessão encerrada.
   - **Desmontar o canvas** corretamente, liberando recursos.
   - **Redirecionar para o login**, substituindo o histórico, de modo que o botão "voltar" não reabra telas protegidas.
3. **Falha de rede** no logout: limpar a sessão local e redirecionar mesmo assim. Registrar o erro, sem telas quebradas.
4. **Várias abas**: ao sair em uma aba, as outras detectam (`storage` event / `BroadcastChannel`) e também vão ao login, **tentando salvar antes** o que estiver pendente.
5. **Sessão expirada (401)**: tratar de forma centralizada. Tentar preservar localmente as alterações pendentes do canvas, levar ao login com "Sua sessão expirou. Entre novamente." e, após novo login, **retornar ao quadro em que estava**.
6. **Rotas protegidas**: sem sessão válida, qualquer rota protegida redireciona ao login; após o login, voltar à página pretendida.
7. **Auditoria**: registrar o evento de logout se já existir log de auditoria.

### 4.3 Caso não exista autenticação
Se a Fase 0 mostrar que **não há login**, **pare** e documente com uma proposta mínima (login simples, sessão segura, guard de rotas) para aprovação. Não invente um sistema completo por conta própria.

---

## 5. Navegação e layout

- Integre Home e menu do usuário ao **layout existente**. Não crie layout paralelo.
- **Início** como primeiro item do menu. Mantenha todos os outros itens como estão.
- Na tela do quadro, adicionar apenas um **atalho discreto "← Início"** na barra superior existente, sem alterar a área do canvas nem suas ferramentas.

---

## 6. Testes e validação

1. **Rode a suíte de testes e o build antes e depois.** Nada que passava pode passar a falhar.
2. Adicione testes **novos** para:
   - Renderização da Home (com dados, vazia, erro por seção).
   - Agregação de etiquetas (contagem correta por etiqueta, respeitando permissões).
   - Proteção de rotas.
   - Logout: salvamento antes de sair, aviso de alterações não salvas, limpeza de estado, redirecionamento, falha de rede, multi-abas, 401.
3. Se não houver framework de testes, **não instale um sem avisar**; entregue um **roteiro de teste manual** passo a passo.

### Checklist de regressão
- [ ] Login continua funcionando.
- [ ] **Quadros antigos abrem idênticos** (cartões, conexões, etiquetas, posições e zoom).
- [ ] Criar, mover, editar e etiquetar cartões continua funcionando no canvas.
- [ ] Execução e orquestração de agentes continuam funcionando.
- [ ] Colaboração em tempo real e presença continuam funcionando.
- [ ] Nenhum contrato de API existente mudou.
- [ ] Nenhum erro novo no console ou nos logs do servidor.
- [ ] Build de produção sem erros nem warnings novos relevantes.
- [ ] Home abre em ~2 s sem carregar o canvas.
- [ ] Sair com alterações pendentes **não perde dados**.
- [ ] Após logout, "voltar" do navegador não reabre quadros.

---

## 7. Definição de pronto

- [ ] Relatório de descoberta em `docs/DESCOBERTA_HOME_LOGOUT.md`.
- [ ] Home com todas as seções aplicáveis e estados (carregando/vazio/erro).
- [ ] Painel de produção usando as **etiquetas reais** do sistema.
- [ ] Logout completo conforme a seção 4, **sem perda de dados do canvas**.
- [ ] Rotas protegidas funcionando.
- [ ] Testes novos passando e **sem regressão** nos antigos.
- [ ] Nenhum dado falso apresentado como real.
- [ ] Relatório final entregue.

---

## 8. Relatório final (obrigatório)

Crie `docs/RELATORIO_HOME_LOGOUT.md` com:

1. Resumo do que foi feito.
2. Lista de arquivos criados e alterados (com o motivo de cada alteração em arquivo existente).
3. Endpoints e migrações novos.
4. Dependências adicionadas e por quê.
5. Resultado dos testes e do build (antes e depois).
6. Decisões e suposições.
7. Pendências e riscos conhecidos.
8. Sugestões de próximos passos para a gestão da empresa (ex.: modelos de quadro, visão por equipe, financeiro, clientes, relatórios de produtividade, permissões por papel, notificações, etiquetas configuráveis por projeto).

---

## 9. Como proceder

Ordem: **Fase 0 (descoberta) → plano curto → implementação em etapas pequenas com commits → testes → relatório final.**

Antes de implementar, apresente um **plano resumido** (arquivos a criar/alterar) e aguarde confirmação se algo for de alto risco. Na dúvida entre "mudar o existente" e "adicionar algo novo ao lado", **sempre adicione**.

