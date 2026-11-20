-- name: QueueFactoryJob :one
INSERT INTO geev.factory_jobs (owner_id, card_id, idempotency_key)
VALUES (auth.uid(), $1, $2)
ON CONFLICT (owner_id, idempotency_key) DO NOTHING
RETURNING id, card_id, status, attempts;

-- name: GetFactoryJobByKey :one
SELECT id, card_id, status, attempts FROM geev.factory_jobs
WHERE owner_id = auth.uid() AND idempotency_key = $1;

-- name: GetFactoryJobByID :one
SELECT id, card_id, status, attempts, lease_expires_at FROM geev.factory_jobs
WHERE owner_id = auth.uid() AND id = $1;

-- Varredura paginada para reenviar sinais perdidos; nunca inclui tarefas em execução.
-- name: ListQueuedFactoryJobIDs :many
SELECT id FROM geev.factory_jobs
WHERE owner_id = auth.uid() AND status = 'queued' AND id > $1
ORDER BY id LIMIT $2;

-- Um processo só assume trabalho pendente; leases expirados exigem análise antes de repetir IA.
-- name: ClaimFactoryJob :one
WITH candidate AS (
    SELECT id FROM geev.factory_jobs
    WHERE owner_id = auth.uid() AND status = 'queued'
    ORDER BY created_at, id FOR UPDATE SKIP LOCKED LIMIT 1
)
UPDATE geev.factory_jobs AS job
SET status = 'running', lease_token = $1, lease_expires_at = clock_timestamp() + interval '5 minutes',
    attempts = attempts + 1, updated_at = clock_timestamp()
FROM candidate WHERE job.id = candidate.id
RETURNING job.id, job.card_id, job.lease_token, job.lease_expires_at, job.attempts;

-- A mensagem identifica exatamente uma tarefa; repetição ou tarefa de outro dono não assume outra.
-- name: ClaimFactoryJobByID :one
UPDATE geev.factory_jobs AS job
SET status = 'running', lease_token = $2, lease_expires_at = clock_timestamp() + interval '5 minutes',
    attempts = attempts + 1, updated_at = clock_timestamp()
WHERE job.id = $1 AND job.owner_id = auth.uid() AND job.status = 'queued'
RETURNING job.id, job.card_id, job.lease_token, job.lease_expires_at, job.attempts;

-- name: FinishFactoryJob :execrows
UPDATE geev.factory_jobs
SET status = $3, lease_token = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
WHERE id = $1 AND owner_id = auth.uid() AND status = 'running' AND lease_token = $2
  AND lease_expires_at > clock_timestamp() AND $3 IN ('completed', 'failed');

-- Preserva pausas e resultados finais alterados concorrentemente.
-- name: MarkActiveFactoryPipelineFailed :exec
UPDATE geev.card_pipelines
SET status = 'failed', current_task = 'A execução não foi concluída. Confira a configuração dos agentes e o histórico antes de tentar novamente.'
WHERE owner_id = auth.uid() AND card_id = $1 AND status IN ('queued', 'running');

-- name: LockValidFactoryJob :one
SELECT id FROM geev.factory_jobs
WHERE id = $1 AND owner_id = auth.uid() AND status = 'running' AND lease_token = $2
  AND lease_expires_at > clock_timestamp()
FOR UPDATE;

-- name: MarkExpiredFactoryJobs :one
WITH expired AS (
    UPDATE geev.factory_jobs AS job
    SET status = 'interrupted', lease_token = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
    WHERE owner_id = auth.uid() AND status = 'running' AND lease_expires_at <= clock_timestamp()
    RETURNING card_id
), pipelines AS (
    UPDATE geev.card_pipelines AS p
    SET status = 'interrupted', current_task = 'Execução interrompida. Revise o histórico e os custos antes de solicitar nova tentativa.'
    FROM expired e
    WHERE p.owner_id = auth.uid() AND p.card_id = e.card_id AND p.status IN ('queued', 'running')
    RETURNING p.card_id
)
SELECT count(*)::bigint FROM expired;

-- name: MarkExpiredFactoryJobForCard :one
WITH expired AS (
    UPDATE geev.factory_jobs AS job
    SET status = 'interrupted', lease_token = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
    WHERE job.owner_id = auth.uid() AND job.card_id = $1 AND job.status = 'running' AND job.lease_expires_at <= clock_timestamp()
    RETURNING job.card_id
), pipelines AS (
    UPDATE geev.card_pipelines AS p
    SET status = 'interrupted', current_task = 'Execução interrompida. Revise o histórico e os custos antes de solicitar nova tentativa.'
    FROM expired e
    WHERE p.owner_id = auth.uid() AND p.card_id = e.card_id AND p.status IN ('queued', 'running')
    RETURNING p.card_id
)
SELECT count(*)::bigint FROM expired;

-- name: PutFactoryWorkspaceFile :exec
INSERT INTO geev.factory_workspace_files (owner_id, card_id, path, content)
VALUES (auth.uid(), $1, $2, $3)
ON CONFLICT (owner_id, card_id, path) DO UPDATE SET content = EXCLUDED.content, updated_at = clock_timestamp();

-- Serializa substituições completas e confirma que o cartão pertence ao dono.
-- name: LockFactoryWorkspace :one
SELECT id FROM geev.cards WHERE owner_id = auth.uid() AND id = $1 FOR UPDATE;

-- name: DeleteFactoryWorkspaceFiles :exec
DELETE FROM geev.factory_workspace_files WHERE owner_id = auth.uid() AND card_id = $1;

-- name: ListFactoryWorkspaceFiles :many
SELECT path, content FROM geev.factory_workspace_files
WHERE owner_id = auth.uid() AND card_id = $1 ORDER BY path;
