-- name: CreatePaymentOrder :exec
INSERT INTO geev.payment_orders (
    id, owner_id, card_id, client_id, amount_cents, currency, pix_emv, pix_txid, pix_key, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);

-- name: GetPaymentOrderByID :one
SELECT id, owner_id, card_id, client_id, amount_cents, currency, pix_emv, pix_txid, pix_key, status, paid_at, created_at, updated_at
FROM geev.payment_orders
WHERE id = $1;

-- name: GetPaymentOrderByTxID :one
SELECT id, owner_id, card_id, client_id, amount_cents, currency, pix_emv, pix_txid, pix_key, status, paid_at, created_at, updated_at
FROM geev.payment_orders
WHERE pix_txid = $1;

-- name: GetPaymentOrderByCardID :one
SELECT id, owner_id, card_id, client_id, amount_cents, currency, pix_emv, pix_txid, pix_key, status, paid_at, created_at, updated_at
FROM geev.payment_orders
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdatePaymentOrderStatus :exec
UPDATE geev.payment_orders
SET status = $2, paid_at = $3, updated_at = $4
WHERE id = $1;
