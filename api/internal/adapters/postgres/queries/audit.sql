-- name: RecordAudit :exec
INSERT INTO geev.audit_log (
    id, owner_id, action, actor_type, actor_id, target_type, target_id, meta, at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: ListAuditLogs :many
SELECT id, owner_id, action, actor_type, actor_id, target_type, target_id, meta, at
FROM geev.audit_log
ORDER BY at DESC
LIMIT $1;
