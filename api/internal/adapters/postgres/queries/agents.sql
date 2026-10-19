-- name: GetAgentByTokenHash :one
SELECT id, owner_id, role, active, spend_limit_cents, spend_current_cents, scope, token_hash, created_at, updated_at
FROM geev.agents
WHERE token_hash = $1
LIMIT 1;

-- name: ListAgents :many
SELECT id, owner_id, role, active, spend_limit_cents, spend_current_cents, scope, created_at, updated_at
FROM geev.agents
ORDER BY role ASC;

-- name: CreateAgent :one
INSERT INTO geev.agents (
    id, owner_id, role, active, spend_limit_cents, spend_current_cents, scope, token_hash, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING id, owner_id, role, active, spend_limit_cents, spend_current_cents, scope, created_at, updated_at;

-- name: UpdateAgentSpend :exec
UPDATE geev.agents
SET spend_current_cents = spend_current_cents + $2,
    updated_at = $3
WHERE id = $1;

-- name: ToggleAgent :exec
UPDATE geev.agents
SET active = $2,
    updated_at = $3
WHERE id = $1;
