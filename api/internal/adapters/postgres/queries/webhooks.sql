-- name: CreateWebhookEndpoint :exec
INSERT INTO geev.webhook_endpoints (
    id, owner_id, url, description, secret_token, events, active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetWebhookEndpointByID :one
SELECT * FROM geev.webhook_endpoints
WHERE id = $1;

-- name: ListWebhookEndpoints :many
SELECT * FROM geev.webhook_endpoints
WHERE ($1::boolean = false OR active = true)
ORDER BY created_at DESC;

-- name: DeleteWebhookEndpoint :exec
DELETE FROM geev.webhook_endpoints
WHERE id = $1;

-- name: CreateWebhookDelivery :exec
INSERT INTO geev.webhook_deliveries (
    id, owner_id, webhook_id, event, payload, status_code, attempts, response_body, success, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: ListWebhookDeliveries :many
SELECT * FROM geev.webhook_deliveries
WHERE ($1::uuid IS NULL OR webhook_id = $1)
ORDER BY created_at DESC
LIMIT $2;

-- name: GetNotificationPreferences :one
SELECT * FROM geev.notification_preferences
WHERE owner_id = $1;

-- name: UpsertNotificationPreferences :exec
INSERT INTO geev.notification_preferences (
    id, owner_id, channels, events, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (owner_id) DO UPDATE
SET channels = EXCLUDED.channels,
    events = EXCLUDED.events,
    updated_at = EXCLUDED.updated_at;
