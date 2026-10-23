-- name: CreateSubscription :exec
INSERT INTO geev.subscriptions (
    id, owner_id, client_id, app_id, plan, description, amount_cents, currency,
    interval, due_day, status, start_date, end_date, next_due_date, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
);

-- name: GetSubscriptionByID :one
SELECT * FROM geev.subscriptions
WHERE id = $1;

-- name: ListSubscriptions :many
SELECT * FROM geev.subscriptions
WHERE ($1::uuid IS NULL OR client_id = $1)
  AND ($2::text IS NULL OR status = $2)
ORDER BY created_at DESC;

-- name: UpdateSubscriptionStatus :exec
UPDATE geev.subscriptions
SET status = $2,
    next_due_date = COALESCE($3, next_due_date),
    updated_at = now()
WHERE id = $1;

-- name: UpdateSubscriptionNextDueDate :exec
UPDATE geev.subscriptions
SET next_due_date = $2,
    updated_at = now()
WHERE id = $1;

-- name: CreateSubscriptionCharge :exec
INSERT INTO geev.subscription_charges (
    id, owner_id, subscription_id, amount_cents, currency, due_date, status, payment_method, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (subscription_id, due_date) DO NOTHING;

-- name: GetSubscriptionChargeByID :one
SELECT * FROM geev.subscription_charges
WHERE id = $1;

-- name: ListSubscriptionCharges :many
SELECT * FROM geev.subscription_charges
WHERE ($1::text IS NULL OR status = $1)
  AND ($2::date IS NULL OR due_date <= $2)
ORDER BY due_date ASC;

-- name: UpdateSubscriptionChargeStatus :exec
UPDATE geev.subscription_charges
SET status = $2,
    paid_at = $3,
    payment_method = COALESCE($4, payment_method),
    updated_at = now()
WHERE id = $1;
