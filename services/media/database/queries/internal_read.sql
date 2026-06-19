-- name: ListMediaByIDs :many
SELECT id, scope, display_name, mime_type, media_kind, object_key
FROM media_assets
WHERE id = ANY($1::uuid[]);
