-- name: ListConnectors :many
SELECT id, owner_id, name, active, secret_ref, scope, created_at, updated_at
FROM geev.connectors
ORDER BY name ASC;

-- name: GetConnectorByID :one
SELECT id, owner_id, name, active, secret_ref, scope, created_at, updated_at
FROM geev.connectors
WHERE id = $1;

-- name: UpsertConnector :one
INSERT INTO geev.connectors (
    id, owner_id, name, active, secret_ref, scope, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (id) DO UPDATE
SET active = EXCLUDED.active,
    secret_ref = EXCLUDED.secret_ref,
    scope = EXCLUDED.scope,
    updated_at = EXCLUDED.updated_at
RETURNING id, owner_id, name, active, secret_ref, scope, created_at, updated_at;
