-- name: GetLanguageIDByCode :one
SELECT id
FROM languages
WHERE code = $1;

-- name: GetLanguageByCode :one
SELECT id, code, name, native_name, direction, is_active, created_at
FROM languages
WHERE code = $1;

-- name: ListActiveLanguages :many
SELECT id, code, name, native_name, direction, is_active, created_at
FROM languages
WHERE is_active = true
ORDER BY code;

-- name: ListAllLanguages :many
SELECT id, code, name, native_name, direction, is_active, created_at
FROM languages
ORDER BY code;

-- name: CreateLanguage :one
INSERT INTO languages (code, name, native_name, direction, is_active)
VALUES ($1, $2, $3, $4, true)
RETURNING id, code, name, native_name, direction, is_active, created_at;

-- name: UpdateLanguage :one
UPDATE languages
SET
    name = COALESCE(sqlc.narg('name'), name),
    native_name = COALESCE(sqlc.narg('native_name'), native_name),
    direction = COALESCE(sqlc.narg('direction'), direction),
    is_active = COALESCE(sqlc.narg('is_active'), is_active)
WHERE code = sqlc.arg('code')
RETURNING id, code, name, native_name, direction, is_active, created_at;
