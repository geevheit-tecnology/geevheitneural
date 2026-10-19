-- name: ListAppsWithFinancials :many
SELECT 
    a.id, 
    a.owner_id, 
    a.client_id, 
    COALESCE(c.name, 'sem cliente') AS client_name,
    a.name, 
    a.hosting, 
    a.database, 
    a.repository, 
    a.domain, 
    a.status,
    a.created_at, 
    a.updated_at,
    COALESCE(
        (SELECT SUM(l.amount_cents) FROM geev.ledger l WHERE l.app_id = a.id AND l.kind = 'custo'), 
        0
    )::bigint AS monthly_cost_cents,
    COALESCE(
        (SELECT SUM(l.amount_cents) FROM geev.ledger l WHERE l.app_id = a.id AND l.kind = 'receita'), 
        0
    )::bigint AS monthly_revenue_cents
FROM geev.apps a
LEFT JOIN geev.clients c ON a.client_id = c.id
ORDER BY a.name ASC;

-- name: GetAppByID :one
SELECT id, owner_id, client_id, name, hosting, database, repository, domain, status, created_at, updated_at
FROM geev.apps
WHERE id = $1;

-- name: CreateApp :one
INSERT INTO geev.apps (
    id, owner_id, client_id, name, hosting, database, repository, domain, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, owner_id, client_id, name, hosting, database, repository, domain, status, created_at, updated_at;

-- name: UpdateApp :exec
UPDATE geev.apps
SET client_id = $2,
    name = $3,
    hosting = $4,
    database = $5,
    repository = $6,
    domain = $7,
    status = $8,
    updated_at = $9
WHERE id = $1;

-- name: DeleteApp :exec
DELETE FROM geev.apps
WHERE id = $1;
