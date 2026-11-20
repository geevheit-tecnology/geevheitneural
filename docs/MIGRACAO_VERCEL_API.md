# Migração da API GeevheitNeural para a Vercel

## Estado verificado em 2026-11-20

- A interface `web` está na Vercel, domínio `geevheit.vercel.app`; o frontend ainda aponta para `geevheit-api.onrender.com`.
- O projeto Supabase existente `tggzeapudgzswdojtndj` permanece responsável pelo banco e autenticação. Não criar nem substituir identidades ou banco.
- Não existe um projeto Vercel específico desta API na lista consultada. Os projetos de outros produtos não devem ser reutilizados.
- O commit `3c81670` foi enviado ao repositório e a publicação Vercel da interface ficou Ready. O estado do deploy correspondente da API no Render não foi comprovado.
- A função da migration 00011 foi aplicada diretamente no Supabase e conferida; a aplicação manual não garante registro da versão no histórico Goose.

## Dependências que impedem uma troca direta

1. `StartBuildUseCase` responde ao pedido e continua a geração em goroutine; a execução precisa sobreviver ao encerramento da função e a novas tentativas sem duplicar cobrança ou arquivos.
2. `LocalSandboxManager` e `LocalPublisher` dependem de arquivos temporários, processo Git e memória locais. As prévias e os arquivos precisam de armazenamento durável e privado, com controle de acesso e limpeza definida.
3. `InMemoryPublisher`, SSE e o consumidor de notificações vivem na memória de um processo. Instâncias independentes não compartilham eventos; substituir por mecanismo persistente ou atualização por consulta e preservar comportamento comprovado.
4. `main.go` sobe um servidor, listeners e pool; a Vercel precisa de um handler Go por requisição, inicialização controlada e conexões compatíveis com serverless.
5. Portões, estado dos agentes, isolamento RLS, autenticação AAL2, limite de gastos e eventos de auditoria precisam manter suas garantias.

## Fases propostas

1. **Preparação local concluída:** `internal/bootstrap.NewHandler` compartilha as rotas com o servidor existente; `api/api/index.go` fornece entrada experimental Go, desabilitada por padrão e limitada a duas conexões por instância. Testes com `go test -race ./...` incluem health e rota protegida. A entrada não foi publicada, nem validada no runtime Vercel ou com banco de teste conectado. Não configurar `GEEV_SERVERLESS_EXPERIMENTAL=1` em produção até as próximas fases.
2. **Base local iniciada:** migration 00012 cria tarefas com idempotência/lease e arquivos de texto no PostgreSQL, com RLS e chave composta dono/cartão; consultas sqlc e repositório foram adicionados. Testes em transação revertida no Supabase local verificaram isolamento, chave repetida, token incorreto, lease expirado e interrupção sem repetição automática. Ainda falta conectar a esteira, implementar recuperação supervisionada, testar concorrência entre processos e validar limites de armazenamento/custo. Migration não aplicada em produção.
3. Adaptar atualização de interface e notificações ao ambiente distribuído, testar falha, reinício, concorrência, pausa e aprovação humana.
4. Publicar API em projeto Vercel separado, configurar segredos apenas no servidor e homologar sessão real, geração, prévia e exclusão. Só depois alterar a URL da interface e confirmar o novo fluxo.
5. Desativar Render somente após validação e autorização do dono, preservando rollback.

Documentação consultada: [runtime Go](https://vercel.com/docs/functions/runtimes/go), [filesystem das funções](https://vercel.com/docs/functions/runtimes), [Vercel Queues](https://vercel.com/docs/queues), [armazenamento privado](https://vercel.com/docs/vercel-blob/private-storage). A escolha final de fila e armazenamento depende de validação do projeto e dos limites do plano.

## Retomada e validação local

Retomar pela integração da base durável com a esteira e a prévia, incluindo o mecanismo que aciona o trabalho no ambiente serverless. As consultas e o repositório novos ainda não são chamados pelo pipeline; a existência das tabelas não torna a execução durável. Evitar qualquer troca da URL do frontend até comprovar esse caminho, a autorização e a entrega real em ambiente isolado.

- Suíte da API: `cd api && go test -race ./...`.
- Teste de integração opt-in: com Supabase local na porta 54332 e banco descartável ou de teste autorizado, definir `GEEV_FACTORY_ROLLBACK_TEST=1` e `GEEV_FACTORY_ROLLBACK_DATABASE_URL` para esse banco e executar `go test -race ./internal/adapters/postgres -run '^TestDurableFactoryPostgresRollback$' -count=1 -v` em `api`. O teste aplica a migration e cria dados sintéticos em transação revertida; requer privilégios para criar esquema e assumir `authenticated`. Nunca apontar para produção.
- `db/tests/factory_durable.sql` complementa o teste do adaptador e deve ser executado após o trecho Up da migration 00012, na mesma transação revertida.
- O código atual foi preparado para uma branch de trabalho; não fazer merge para `main` enquanto o caminho serverless, a migration e as integrações não estiverem homologados. Push da branch pode criar uma prévia automática; não equivale a publicação em produção.
