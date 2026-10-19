-- name: ListClients :many
SELECT id, owner_id, name, contact_ref, created_at, updated_at
FROM geev.clients
ORDER BY name ASC;

-- name: GetClientByID :one
SELECT id, owner_id, name, contact_ref, created_at, updated_at
FROM geev.clients
WHERE id = $1;

-- name: FindClientByName :one
SELECT id, owner_id, name, contact_ref, created_at, updated_at
FROM geev.clients
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: CreateClient :one
INSERT INTO geev.clients (
    id, owner_id, name, contact_ref, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, owner_id, name, contact_ref, created_at, updated_at;

-- name: DeleteClient :exec
DELETE FROM geev.clients
WHERE id = $1;
