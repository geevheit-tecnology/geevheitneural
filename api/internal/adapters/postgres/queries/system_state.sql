-- name: GetSystemState :one
SELECT id, owner_id, paused, paused_by, paused_at, reason, updated_at
FROM geev.system_state
WHERE id = true;

-- name: SaveSystemState :exec
UPDATE geev.system_state
SET paused = $1,
    paused_by = $2,
    paused_at = $3,
    reason = $4,
    updated_at = $5
WHERE id = true;
