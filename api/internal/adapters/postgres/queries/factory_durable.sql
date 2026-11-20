-- name: QueueFactoryJob :one
INSERT INTO geev.factory_jobs (owner_id, card_id, idempotency_key)
VALUES (auth.uid(), $1, $2)
ON CONFLICT (owner_id, idempotency_key) DO NOTHING
RETURNING id, card_id, status, attempts;

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

-- name: FinishFactoryJob :execrows
UPDATE geev.factory_jobs
SET status = $3, lease_token = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
WHERE id = $1 AND owner_id = auth.uid() AND status = 'running' AND lease_token = $2
  AND lease_expires_at > clock_timestamp() AND $3 IN ('completed', 'failed');

-- name: MarkExpiredFactoryJobs :execrows
UPDATE geev.factory_jobs
SET status = 'interrupted', lease_token = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
WHERE owner_id = auth.uid() AND status = 'running' AND lease_expires_at <= clock_timestamp();

-- name: PutFactoryWorkspaceFile :exec
INSERT INTO geev.factory_workspace_files (owner_id, card_id, path, content)
VALUES (auth.uid(), $1, $2, $3)
ON CONFLICT (owner_id, card_id, path) DO UPDATE SET content = EXCLUDED.content, updated_at = clock_timestamp();

-- name: ListFactoryWorkspaceFiles :many
SELECT path, content FROM geev.factory_workspace_files
WHERE owner_id = auth.uid() AND card_id = $1 ORDER BY path;
