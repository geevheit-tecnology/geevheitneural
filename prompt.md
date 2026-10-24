# GeevheitNeural — Prompt Mestre & Plano de Execução da Fábrica Autônoma (100% Gratuita)

> **Documento Guia Operacional**: Este arquivo serve como bússola e memória técnica para orientar os agentes de IA na implementação e operação do pipeline autônomo de geração de software sem custos de terceiros.
> **Status:** Planejamento e Arquitetura Inicial
> **Referência:** `docs/MEMORY.md`, `docs/ARCHITECTURE.md`, `docs/WEBHOOKS_AND_ORCHESTRATION.md`

---

## 1. Objetivo Principal

Transformar o **GeevheitNeural** em uma **fábrica autônoma de software** capaz de conceber, arquitetar, codificar, testar e publicar aplicações web reais (landing pages, CRUDs, PWAs, APIs) a partir de uma ordem do Dono ou pedido de cliente, operando **100% dentro de camadas gratuitas (R$ 0,00)** e mantendo o Dono no comando absoluto através de portões de aprovação 2FA.

---

## 2. Invariantes Invioláveis de Arquitetura

1. **Portões Exclusivos do Dono:**
   - Nenhum agente avança portão (`proposta`, `prompt`, `entrega`).
   - Todo avanço de portão exige `domain.OwnerPrincipal` autenticado com 2FA (`aal2`).
2. **`ExecutionGuard` Mandatório:**
   - Se o sistema estiver pausado (`system.is_paused == true`) ou o conector estiver desativado (`ativo == false`), nenhuma execução de LLM ou chamada de ferramenta é permitida.
3. **Custo Zero (R$ 0,00) de Início:**
   - Utilização estrita de serviços com Free Tier sem necessidade de pagamento antecipado:
     - **LLM:** Google Gemini API (Google AI Studio Free Tier — até 1.500 req/dia) ou Ollama local (`qwen2.5-coder:7b`).
     - **Sandbox:** Docker local temporário ou processos isolados em subdiretório de trabalho efêmero (`workspace/scratch/<card-id>`).
     - **Controle de Versão:** GitHub API (plano gratuito).
     - **Deploy de Homologação:** Vercel CLI / Cloudflare Pages (plano Hobby gratuito).
4. **Clean Architecture e SOLID no Backend Go:**
   - Camadas: `domain` (puro) → `app` (casos de uso) → `adapters` (Gemini, Git, Sandbox, Vercel) → `cmd` (injeção em `main.go`).
   - Implementação de `pipeline.AgentRunner` substituindo o mock sem quebrar contratos.
5. **Segredos e Credenciais:**
   - Nenhuma API Key ou token no código ou no banco. Todas as credenciais via `secret_ref` e variáveis de ambiente locais/servidor.

---

## 3. Topologia do Pipeline de Agentes

```
                                 [Ordem do Dono ou Intake]
                                             │
                                             ▼
                                  [1. Agente Requisitos]
                            (Analisa escopo e cria proposta BRL)
                                             │
                                             ▼
                                  [PORTÃO 1: Proposta] ◄─── DONO APROVA (2FA)
                                             │
                                             ▼
                                    [2. Agente Arquiteto]
                               (Gera plano técnico e tarefas)
                                             │
                                             ▼
                                    [PORTÃO 2: Prompt]   ◄─── DONO APROVA (2FA)
                                             │
                                             ▼
                                      [3. Agente Dev]
                             (Gera código no workspace isolado)
                                             │
                                             ▼
                                       [4. Agente QA]
                              (Roda testes locais e valida build)
                                             │
                                             ▼
                                    [5. Agente Segurança]
                             (Varre segredos e vulnerabilidades)
                                             │
                                             ▼
                                    [6. Agente Publicador]
                               (Deploy preview Vercel / GitHub)
                                             │
                                             ▼
                                  [PORTÃO 3: Entrega]   ◄─── DONO APROVA (2FA)
                                             │
                                             ▼
                                  [7. Portal do Cliente]
                               (Aceite formal com SHA-256)
```

---

## 4. Matriz de Ferramentas e Custos

| Papel | Provedor / Tecnologia Gratuita | Mecanismo de Execução | Custo |
|---|---|---|---|
| **Requisitos & Arquiteto** | Google Gemini 1.5/2.0 Flash (AI Studio) | Chamada HTTP direta com System Prompt BDD | R$ 0,00 |
| **Desenvolvedor (Dev)** | Gemini 1.5/2.0 Flash ou Ollama Local | Geração de arquivos multi-file em sandbox | R$ 0,00 |
| **QA (Testes)** | Runner Local (Go/Node) / GitHub Actions | Execução de linters e testes unitários | R$ 0,00 |
| **Segurança** | Scripting Go / Gitleaks local | Análise estática do código gerado | R$ 0,00 |
| **Publicador** | Vercel CLI / Cloudflare Pages API | Publicação automática do preview | R$ 0,00 |
| **Repositório** | GitHub REST API | Criação de repositório e branch remota | R$ 0,00 |

---

## 5. Roteiro de Implementação (Passo a Passo)

### Etapa 1: Adapter do Google Gemini (`GeminiAgentRunner`)
- [x] Criar `api/internal/adapters/agents/gemini_runner.go`.
- [x] Implementar a interface `pipeline.AgentRunner`.
- [x] Ler `GEMINI_API_KEY` do ambiente via `secret_ref`.
- [x] Adicionar rate limiter interno (máximo 15 requisições/minuto) para evitar erro `429 Too Many Requests`.
- [x] Testes unitários do runner com fakes de resposta HTTP.

### Etapa 2: Gerenciador de Sandbox Local (`LocalSandboxManager`)
- [x] Criar diretório isolado de trabalho efêmero por projeto: `scratch/builds/<card-id>/`.
- [x] Implementar gerador de arquivos (criação da árvore de arquivos a partir da resposta JSON/Markdown do LLM).
- [x] Executar comandos de validação local de forma segura (`npm install`, `npm run build` ou `go build`).

### Etapa 3: Prompts Especializados por Papel
- [x] **Prompt Requisitos:** Formata escopo, histórias de usuário BDD (Given-When-Then) e estimativa de preço em centavos.
- [x] **Prompt Arquiteto:** Define stack (ex: HTML/Tailwind/JS ou Next.js static), estrutura de pastas e lista de arquivos necessários.
- [x] **Prompt Dev:** Devolve código limpo e completo (sem placeholders `// TODO` ou código truncado).
- [x] **Prompt QA:** Analisa o código gerado em busca de erros lógicos, acessibilidade e conformidade com o escopo.

### Etapa 4: Conector GitHub & Deploy de Prévia
- [x] Criar adapter simples de publicação (`LocalPublisher`):
  - Inicializar Git local no diretório do projeto.
  - Commits atômicos padronizados.
  - Endpoints de prévia e listagem de workspace: `GET /v1/cards/{id}/preview` e `GET /v1/cards/{id}/workspace`.

### Etapa 5: Injeção no `cmd/geevheit-api/main.go` e Teste de Fluxo Real
- [x] Configurar chave seletora (`AGENT_RUNNER_TYPE=gemini|mock` com fallback automático).
- [x] Integração limpa no `StartBuildUseCase` com gravação física na sandbox e publicação atômica.
- [x] Validação completa de testes com `-race` (100% verde).
- [x] Custo zero absoluto mantido (R$ 0,00).


---

## 6. Armadilhas Conhecidas e Mitigações

| Armadilha | Mitigação |
|---|---|
| **Limite de Taxa (Gemini Free 15 RPM)** | Aplicar backoff exponencial e fila com delay mínimo de 4 segundos entre chamadas de agentes consecutivos. |
| **Código Truncado pelo LLM** | Exigir respostas estruturadas por arquivo e limitar o tamanho de cada arquivo gerado ou usar streaming com montagem de buffer. |
| **Comandos Perigosos na Sandbox** | Proibir execução de comandos arbitrários fora da pasta temporária do projeto. Validar caminhos com `filepath.Clean`. |
| **Exposição de Chaves de API** | Utilizar variáveis de ambiente ou o cofre de credenciais já modelado no GeevheitNeural. |

---

## 7. Critérios de Sucesso

1. O Dono cria um pedido de sistema pelo painel ou voz.
2. O sistema gera proposta e plano arquitetural com custo R$ 0,00.
3. O Dono aprova os portões pelo celular.
4. O código é gerado na máquina, testado e publicado em uma URL de preview.
5. O link do Portal do Cliente é liberado para aprovação final.
