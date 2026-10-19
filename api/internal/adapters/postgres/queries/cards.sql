-- name: GetCardByID :one
SELECT id, owner_id, client_id, app_id, type, title, price_cents, est_cost_cents, currency, stage, priority, due_date, created_at, updated_at
FROM geev.cards
WHERE id = $1;

-- name: GetCardEnrichedByID :one
SELECT 
    c.id, 
    c.owner_id, 
    c.client_id, 
    cl.name AS client_name,
    c.app_id, 
    COALESCE(a.name, 'Novo') AS app_name,
    c.type, 
    c.title, 
    c.price_cents, 
    c.est_cost_cents, 
    c.currency, 
    c.stage, 
    c.priority, 
    c.due_date, 
    c.created_at, 
    c.updated_at
FROM geev.cards c
JOIN geev.clients cl ON c.client_id = cl.id
LEFT JOIN geev.apps a ON c.app_id = a.id
WHERE c.id = $1;

-- name: UpdateCard :exec
UPDATE geev.cards
SET stage = $2,
    priority = $3,
    price_cents = $4,
    est_cost_cents = $5,
    updated_at = $6
WHERE id = $1;

-- name: GetMoneyWaiting :one
SELECT COALESCE(SUM(price_cents), 0)::bigint AS total_cents, COUNT(*)::bigint AS count
FROM geev.cards
WHERE stage IN ('proposta_aprovar', 'prompt_aprovar', 'entrega_aprovar');

-- name: ListCards :many
SELECT id, owner_id, client_id, app_id, type, title, price_cents, est_cost_cents, currency, stage, priority, due_date, created_at, updated_at
FROM geev.cards
ORDER BY created_at DESC;

-- name: ListCardsEnriched :many
SELECT 
    c.id, 
    c.owner_id, 
    c.client_id, 
    cl.name AS client_name,
    c.app_id, 
    COALESCE(a.name, 'Novo') AS app_name,
    c.type, 
    c.title, 
    c.price_cents, 
    c.est_cost_cents, 
    c.currency, 
    c.stage, 
    c.priority, 
    c.due_date, 
    c.created_at, 
    c.updated_at
FROM geev.cards c
JOIN geev.clients cl ON c.client_id = cl.id
LEFT JOIN geev.apps a ON c.app_id = a.id
ORDER BY c.created_at DESC;

-- name: CreateCard :exec
INSERT INTO geev.cards (
    id, owner_id, client_id, app_id, type, title, price_cents, est_cost_cents, currency, stage, priority, due_date, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);

-- name: ListCardEventsByCardID :many
SELECT id, owner_id, card_id, kind, payload, actor_type, actor_id, at
FROM geev.card_events
WHERE card_id = $1
ORDER BY at ASC;

-- name: GetRiskReportByCardID :one
SELECT id, owner_id, card_id, tests, secrets_scan, outdated_deps, lgpd_points, source, created_at, updated_at
FROM geev.risk_reports
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpsertRiskReport :one
INSERT INTO geev.risk_reports (
    id, owner_id, card_id, tests, secrets_scan, outdated_deps, lgpd_points, source, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) ON CONFLICT (id) DO UPDATE
SET tests = EXCLUDED.tests,
    secrets_scan = EXCLUDED.secrets_scan,
    outdated_deps = EXCLUDED.outdated_deps,
    lgpd_points = EXCLUDED.lgpd_points,
    source = EXCLUDED.source,
    updated_at = EXCLUDED.updated_at
RETURNING id, owner_id, card_id, tests, secrets_scan, outdated_deps, lgpd_points, source, created_at, updated_at;
