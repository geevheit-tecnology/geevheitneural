-- name: CreateCardEvent :exec
INSERT INTO geev.card_events (
    id, owner_id, card_id, kind, payload, actor_type, actor_id, at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: ListCardEventsByCard :many
SELECT id, owner_id, card_id, kind, payload, actor_type, actor_id, at
FROM geev.card_events
WHERE card_id = $1
ORDER BY at ASC;
