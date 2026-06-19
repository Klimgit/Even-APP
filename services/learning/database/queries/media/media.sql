-- name: ListMediaByIDs :many
SELECT id, scope, display_name, mime_type, media_kind
FROM media_assets
WHERE id = ANY($1::uuid[])
  AND mime_type != 'application/octet-stream'
  AND (expires_at IS NULL OR expires_at > now());
