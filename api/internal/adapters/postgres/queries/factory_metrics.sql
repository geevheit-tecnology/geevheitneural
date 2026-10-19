-- name: CreateFactoryMetric :one
INSERT INTO geev.factory_metrics (
    id, owner_id, card_id, lead_time_seconds, ai_cost_cents, infra_cost_cents, first_time_approval, post_delivery_defects, owner_decision_time_seconds, margin_cents, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, owner_id, card_id, lead_time_seconds, ai_cost_cents, infra_cost_cents, first_time_approval, post_delivery_defects, owner_decision_time_seconds, margin_cents, created_at;

-- name: GetFactoryScoreboard :one
SELECT 
    COALESCE(AVG(lead_time_seconds), 0)::bigint AS avg_lead_time_seconds,
    COALESCE(SUM(ai_cost_cents), 0)::bigint AS total_ai_cost_cents,
    COALESCE(SUM(infra_cost_cents), 0)::bigint AS total_infra_cost_cents,
    COALESCE(SUM(margin_cents), 0)::bigint AS total_margin_cents,
    COALESCE(AVG(owner_decision_time_seconds), 0)::bigint AS avg_decision_time_seconds,
    COALESCE(COUNT(*), 0)::bigint AS total_projects,
    COALESCE(SUM(CASE WHEN first_time_approval THEN 1 ELSE 0 END), 0)::bigint AS first_time_approvals,
    COALESCE(SUM(post_delivery_defects), 0)::bigint AS total_defects
FROM geev.factory_metrics;

-- name: CreatePreview :one
INSERT INTO geev.previews (
    id, owner_id, card_id, preview_url, status, expires_at, rollback_plan, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING id, owner_id, card_id, preview_url, status, expires_at, rollback_plan, created_at, updated_at;

-- name: GetPreviewByCardID :one
SELECT id, owner_id, card_id, preview_url, status, expires_at, rollback_plan, created_at, updated_at
FROM geev.previews
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpsertProjectMemory :exec
INSERT INTO geev.project_memory (
    id, owner_id, card_id, client_id, key, category, content, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: ListProjectMemoryByCardID :many
SELECT id, owner_id, card_id, client_id, key, category, content, created_at
FROM geev.project_memory
WHERE card_id = $1
ORDER BY created_at ASC;
