-- name: CreateProposal :one
INSERT INTO geev.proposals (
    id, owner_id, card_id, amount_cents, currency, body, accept_token_hash, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING id, owner_id, card_id, amount_cents, currency, body, accept_token_hash, status, accepted_at, created_at, updated_at;

-- name: GetProposalByTokenHash :one
SELECT 
    p.id, 
    p.owner_id, 
    p.card_id, 
    p.amount_cents, 
    p.currency, 
    p.body, 
    p.accept_token_hash, 
    p.status, 
    p.accepted_at, 
    p.created_at, 
    p.updated_at,
    c.title AS card_title,
    cl.name AS client_name
FROM geev.proposals p
JOIN geev.cards c ON p.card_id = c.id
JOIN geev.clients cl ON c.client_id = cl.id
WHERE p.accept_token_hash = $1;

-- name: AcceptProposal :exec
UPDATE geev.proposals
SET status = 'aceita',
    accepted_at = $2,
    updated_at = $2
WHERE id = $1;

-- name: GetProposalByCardID :one
SELECT id, owner_id, card_id, amount_cents, currency, body, accept_token_hash, status, accepted_at, created_at, updated_at
FROM geev.proposals
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT 1;
