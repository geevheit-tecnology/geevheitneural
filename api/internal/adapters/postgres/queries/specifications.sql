-- name: UpsertSpecification :one
INSERT INTO geev.specifications (
    id, owner_id, card_id, user_stories, tests_code, tests_passed, client_approved, client_approved_at, payment_signal_cents, payment_signal_paid, scope_token_hash, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) ON CONFLICT (card_id) DO UPDATE
SET user_stories = EXCLUDED.user_stories,
    tests_code = EXCLUDED.tests_code,
    tests_passed = EXCLUDED.tests_passed,
    client_approved = EXCLUDED.client_approved,
    client_approved_at = EXCLUDED.client_approved_at,
    payment_signal_cents = EXCLUDED.payment_signal_cents,
    payment_signal_paid = EXCLUDED.payment_signal_paid,
    scope_token_hash = EXCLUDED.scope_token_hash,
    updated_at = EXCLUDED.updated_at
RETURNING id, owner_id, card_id, user_stories, tests_code, tests_passed, client_approved, client_approved_at, payment_signal_cents, payment_signal_paid, scope_token_hash, created_at, updated_at;

-- name: GetSpecificationByCardID :one
SELECT id, owner_id, card_id, user_stories, tests_code, tests_passed, client_approved, client_approved_at, payment_signal_cents, payment_signal_paid, scope_token_hash, created_at, updated_at
FROM geev.specifications
WHERE card_id = $1
LIMIT 1;

-- name: GetSpecificationByScopeTokenHash :one
SELECT id, owner_id, card_id, user_stories, tests_code, tests_passed, client_approved, client_approved_at, payment_signal_cents, payment_signal_paid, scope_token_hash, created_at, updated_at
FROM geev.specifications
WHERE scope_token_hash = $1
LIMIT 1;

-- name: ApproveSpecificationByClient :exec
UPDATE geev.specifications
SET client_approved = true,
    client_approved_at = $2,
    updated_at = $2
WHERE scope_token_hash = $1;
