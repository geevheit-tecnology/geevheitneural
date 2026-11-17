# Correção 006 — diagnóstico e validação local

Data: 2026-11-16. Escopo: `correcao006.md`, reparo mínimo no working tree.

## Problema e causa comprovada

Antes de cada papel, `StartBuildUseCase.runFactoryPipeline` ignorava erros de `RunInTx` e `SystemState.Get`. Repositório/estado ausentes também deixavam `isPaused=false`. Uma falha de banco durante a verificação permitia executar IA sem confirmar a pausa de emergência.

Regressão executada antes da correção: os quatro cenários falharam com `state check bypassed: calls=1`. Isso comprova o defeito local; não identifica a causa de um incidente específico em produção, pois nenhum log atual de incidente foi fornecido ou encontrado entre os arquivos versionados.

## Alterações

- `api/internal/app/pipeline_usecases.go`: verificar erro transacional, erro de leitura e ausência do repositório/estado; chamar o tratamento de falha existente e retornar antes do runner. Caminho de pausa confirmada preservado.
- `api/internal/app/correcao004_test.go`: acrescentar regressão com quatro cenários, exigindo zero chamadas ao runner e motivo da falha preservado.
- `docs/TASKS.md`, `docs/MEMORY.md`, `docs/CHANGELOG.md` e este relatório: evidências e pendências.

Risco: uma indisponibilidade transitória na leitura passa a interromper o build em vez de prosseguir sem permissão confirmada. Se o banco permanecer indisponível, também pode impedir gravar o status de falha; o runner não será chamado. Nenhuma arquitetura, contrato, configuração, dado ou dependência foi alterada.

## Execução e componentes existentes

- API: `make api-run` / `cd api && go run ./cmd/geevheit-api`; lê `.env`, inicializa pool PostgreSQL, autenticação Supabase/JWKS, repositórios, sandbox, publicador Git e runner selecionado no composition root. Requer banco/esquema existentes e configuração do provedor para geração real.
- Frontend: `cd web && npm run dev`; build estático: `npm run build`. Dependências já presentes foram reutilizadas.
- `Makefile`, README, documentação, regras, `prompt01.md` e fluxo indexado pelo CodeGraph inspecionados. Nenhum serviço com acesso a dados reais foi iniciado para diagnóstico.

## Verificações e saída real resumida

Antes da correção, `go test ./internal/app -run TestCorrection006 -count=1`:

```text
--- FAIL: TestCorrection006_StateCheckFailureStopsAgents
transaction_error: state check bypassed: calls=1
read_error: state check bypassed: calls=1
missing_repository: state check bypassed: calls=1
missing_state: state check bypassed: calls=1
FAIL
```

Depois, `go test -race ./internal/app -run TestCorrection006 -count=1 -v`:

```text
--- PASS: TestCorrection006_StateCheckFailureStopsAgents (0.01s)
    --- PASS: TestCorrection006_StateCheckFailureStopsAgents/transaction_error
    --- PASS: TestCorrection006_StateCheckFailureStopsAgents/read_error
    --- PASS: TestCorrection006_StateCheckFailureStopsAgents/missing_repository
    --- PASS: TestCorrection006_StateCheckFailureStopsAgents/missing_state
PASS
ok github.com/geevheit/neuralcenter/api/internal/app 1.901s
```

- `go vet ./... && go test -race ./... && go build ./...`: exit 0; pacote app `ok ... 2.223s`. Testes usam fakes/transportes locais; não equivalem a homologação remota. A primeira execução no sandbox não acessou o cache Go; repetida com permissão. Uma execução anterior coincidiu com a edição do teste e falhou ao importar `sync/atomic`; a execução final sequencial passou.
- `node web/scripts/test-assistant.cjs`: `tests 12`, `pass 12`, `fail 0`.
- `npm run build`: primeira execução interrompida no sandbox (`SecItemCopyMatching failed -50`). Repetição autorizada fora dele: `Compiled successfully`, TypeScript concluído, `Generating static pages (9/9)`; exit 1 com `SecurityFatalError: Your version of Next.js is affected by a known security vulnerability.` Não é build aprovado. Atualização de dependências requer autorização pelo documento da tarefa.
- `git diff --check`: exit 0.

## Pendências e limites

Funcionamento da interrupção confirmado em regressões locais, sem IA remota. Login/microfone físico, geração completa real, configurações e logs atuais de Render/Supabase/Vercel não foram homologados. Hipóteses externas (pausa do Supabase, IPv6, cold start, variáveis remotas, JWT e CORS) não foram confirmadas nem descartadas por testes locais.

A inspeção também encontrou ausência de timeout no fetch do service worker e chave vazia passada ao handler PIX no composition root; não foram modificados ou declarados causas do problema do orquestrador. Merecem diagnóstico separado, em especial a autenticação do webhook antes de uso real.

Não houve atualização/instalação de dependências, alteração de `.env`, banco/dados, commit, push ou deploy. `correcao006.md` já estava não rastreado e foi preservado.
