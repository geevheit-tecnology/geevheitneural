-- name: CreatePortalToken :exec
INSERT INTO geev.client_portal_tokens (
    id, owner_id, card_id, client_id, token_hash, expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetPortalTokenByHash :one
SELECT id, owner_id, card_id, client_id, token_hash, expires_at, revoked_at, last_accessed_at, accepted_at, accepted_by, accepted_ip, accepted_user_agent, accepted_state_hash, created_at, updated_at
FROM geev.client_portal_tokens
WHERE token_hash = $1;

-- name: GetPortalTokenByCardID :one
SELECT id, owner_id, card_id, client_id, token_hash, expires_at, revoked_at, last_accessed_at, accepted_at, accepted_by, accepted_ip, accepted_user_agent, accepted_state_hash, created_at, updated_at
FROM geev.client_portal_tokens
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdatePortalToken :exec
UPDATE geev.client_portal_tokens
SET revoked_at = $2, last_accessed_at = $3, accepted_at = $4, accepted_by = $5, accepted_ip = $6, accepted_user_agent = $7, accepted_state_hash = $8, updated_at = $9
WHERE id = $1;

-- name: RevokePortalTokensByCardID :exec
UPDATE geev.client_portal_tokens
SET revoked_at = $2, updated_at = $2
WHERE card_id = $1 AND revoked_at IS NULL;
