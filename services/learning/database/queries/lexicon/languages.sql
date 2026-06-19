-- name: GetLanguageByID :one
SELECT id, code, name, native_name, direction, is_active, created_at
FROM languages
WHERE id = $1;

-- name: ListLanguagesByIDs :many
SELECT id, code, name, native_name, direction, is_active, created_at
FROM languages
WHERE id = ANY($1::uuid[]);
