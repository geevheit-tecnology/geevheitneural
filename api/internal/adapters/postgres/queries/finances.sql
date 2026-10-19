-- name: GetFinanceTotals :one
SELECT 
    COALESCE(SUM(CASE WHEN kind = 'receita' THEN amount_cents ELSE 0 END), 0)::bigint AS total_revenue_cents,
    COALESCE(SUM(CASE WHEN kind = 'custo' THEN amount_cents ELSE 0 END), 0)::bigint AS total_cost_cents
FROM geev.ledger;

-- name: ListCostByService :many
SELECT 
    COALESCE(service, 'Outros') AS service_name,
    SUM(amount_cents)::bigint AS cost_cents
FROM geev.ledger
WHERE kind = 'custo'
GROUP BY service
ORDER BY cost_cents DESC;

-- name: ListLedgerEntries :many
SELECT 
    l.id,
    l.owner_id,
    l.app_id,
    COALESCE(a.name, 'Geral') AS app_name,
    l.client_id,
    COALESCE(c.name, 'sem cliente') AS client_name,
    l.kind,
    l.amount_cents,
    l.currency,
    l.occurred_on,
    l.service,
    l.description,
    l.created_at
FROM geev.ledger l
LEFT JOIN geev.apps a ON l.app_id = a.id
LEFT JOIN geev.clients c ON l.client_id = c.id
ORDER BY l.occurred_on DESC, l.created_at DESC
LIMIT $1;

-- name: CreateLedgerEntry :one
INSERT INTO geev.ledger (
    id, owner_id, app_id, client_id, kind, amount_cents, currency, occurred_on, service, description, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, owner_id, app_id, client_id, kind, amount_cents, currency, occurred_on, service, description, created_at;
