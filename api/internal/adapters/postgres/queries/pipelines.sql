-- name: UpsertCardPipeline :one
INSERT INTO geev.card_pipelines (
    id, owner_id, card_id, stage, active_role, spend_total_cents, spend_limit_cents, status, current_task, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) ON CONFLICT (card_id) DO UPDATE
SET stage = EXCLUDED.stage,
    active_role = EXCLUDED.active_role,
    spend_total_cents = EXCLUDED.spend_total_cents,
    spend_limit_cents = EXCLUDED.spend_limit_cents,
    status = EXCLUDED.status,
    current_task = EXCLUDED.current_task,
    updated_at = EXCLUDED.updated_at
RETURNING id, owner_id, card_id, stage, active_role, spend_total_cents, spend_limit_cents, status, current_task, created_at, updated_at;

-- name: GetCardPipelineByCardID :one
SELECT id, owner_id, card_id, stage, active_role, spend_total_cents, spend_limit_cents, status, current_task, created_at, updated_at
FROM geev.card_pipelines
WHERE card_id = $1
LIMIT 1;

-- name: ListCardPipelines :many
SELECT id, owner_id, card_id, stage, active_role, spend_total_cents, spend_limit_cents, status, current_task, created_at, updated_at
FROM geev.card_pipelines
ORDER BY updated_at DESC;

-- name: AddPipelineSpend :exec
UPDATE geev.card_pipelines
SET spend_total_cents = spend_total_cents + $2,
    updated_at = $3
WHERE card_id = $1;
