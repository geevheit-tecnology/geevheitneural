-- name: CreateApproval :exec
INSERT INTO geev.approvals (
    id, owner_id, card_id, gate, decision, decided_by, decided_at, note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: ListApprovalsByCard :many
SELECT id, owner_id, card_id, gate, decision, decided_by, decided_at, note
FROM geev.approvals
WHERE card_id = $1
ORDER BY decided_at DESC;
