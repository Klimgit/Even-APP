-- name: GetMediaSizeByObjectKey :one
SELECT size_bytes
FROM media_assets
WHERE object_key = $1
  AND scope = 'platform';

-- name: InsertPendingMedia :exec
INSERT INTO media_assets (
    id, scope, language_id, object_key, bucket, mime_type, media_kind,
    size_bytes, display_name, uploaded_by
) VALUES (
    $1, 'platform', $2, $3, $4, 'application/octet-stream', 'image', $5, '(uploading)', $6
);

-- name: SumActiveMediaSizeByUser :one
SELECT COALESCE(SUM(size_bytes), 0)::bigint
FROM media_assets
WHERE uploaded_by = $1
  AND (expires_at IS NULL OR expires_at > now());

-- name: ConfirmMedia :one
UPDATE media_assets
SET mime_type = $2,
    media_kind = $3,
    size_bytes = $4,
    width = $5,
    height = $6,
    duration_ms = $7,
    display_name = $8,
    linked_lexeme_id = $9,
    language_id = $10,
    bucket = $11,
    expires_at = $12,
    updated_at = now()
WHERE object_key = $1
  AND scope = 'platform'
RETURNING id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
    size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
    uploaded_by, expires_at, created_at;

-- name: CountPlatformMedia :one
SELECT COUNT(*)::int
FROM media_assets
WHERE scope = 'platform'
  AND language_id = $1
  AND mime_type != 'application/octet-stream'
  AND (expires_at IS NULL OR expires_at > now())
  AND (sqlc.narg('media_kind')::text IS NULL OR media_kind = sqlc.narg('media_kind'))
  AND (sqlc.narg('search')::text IS NULL OR display_name ILIKE sqlc.narg('search'));

-- name: ListPlatformMedia :many
SELECT id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
    size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
    uploaded_by, expires_at, created_at
FROM media_assets
WHERE scope = 'platform'
  AND language_id = $1
  AND mime_type != 'application/octet-stream'
  AND (expires_at IS NULL OR expires_at > now())
  AND (sqlc.narg('media_kind')::text IS NULL OR media_kind = sqlc.narg('media_kind'))
  AND (sqlc.narg('search')::text IS NULL OR display_name ILIKE sqlc.narg('search'))
ORDER BY created_at DESC
LIMIT sqlc.arg('row_limit') OFFSET sqlc.arg('row_offset');

-- name: GetMediaByID :one
SELECT id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
    size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
    uploaded_by, expires_at, created_at
FROM media_assets
WHERE id = $1
  AND scope = 'platform';

-- name: UpdateMedia :one
UPDATE media_assets
SET display_name = $2,
    linked_lexeme_id = $3,
    expires_at = $4,
    updated_at = now()
WHERE id = $1
RETURNING id, scope, language_id, owner_id, object_key, bucket, mime_type, media_kind,
    size_bytes, width, height, duration_ms, display_name, linked_lexeme_id,
    uploaded_by, expires_at, created_at;

-- name: DeleteMedia :exec
DELETE FROM media_assets
WHERE id = $1
  AND scope = 'platform';
