-- name: InsertAuditEvent :exec
INSERT INTO audit_events (actor_id, action, target_type, target_id, details)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditEvents :many
SELECT id, actor_id, action, target_type, target_id, details, created_at
FROM audit_events
WHERE ($1::text IS NULL OR $1 = '' OR action = $1)
ORDER BY created_at DESC
LIMIT $3 OFFSET $2;

-- name: CountAuditEvents :one
SELECT count(*)::int AS count
FROM audit_events
WHERE ($1::text IS NULL OR $1 = '' OR action = $1);
