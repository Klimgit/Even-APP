-- name: ListCoursesByOwner :many
SELECT id, title, target_language_id, ui_language_id, owner_id, is_published, visibility, created_at, updated_at
FROM courses
WHERE owner_id = $1
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR title ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY updated_at DESC, title;

-- name: ListAllCourses :many
SELECT id, title, target_language_id, ui_language_id, owner_id, is_published, visibility, created_at, updated_at
FROM courses
WHERE (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR title ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY updated_at DESC, title
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAllCourses :one
SELECT COUNT(*)::int AS count
FROM courses
WHERE (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR title ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: GetCourseByID :one
SELECT id, title, target_language_id, ui_language_id, owner_id, is_published, visibility, created_at, updated_at
FROM courses
WHERE id = $1;

-- name: GetCourseOwner :one
SELECT owner_id
FROM courses
WHERE id = $1;

-- name: CreateCourse :one
INSERT INTO courses (title, target_language_id, ui_language_id, owner_id, visibility)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, target_language_id, ui_language_id, owner_id, is_published, visibility, created_at, updated_at;

-- name: UpdateCourse :one
UPDATE courses
SET
    title = COALESCE(sqlc.narg('title'), title),
    target_language_id = COALESCE(sqlc.narg('target_language_id'), target_language_id),
    ui_language_id = COALESCE(sqlc.narg('ui_language_id'), ui_language_id),
    visibility = COALESCE(sqlc.narg('visibility'), visibility),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, title, target_language_id, ui_language_id, owner_id, is_published, visibility, created_at, updated_at;

-- name: PublishCourse :one
UPDATE courses
SET is_published = true, updated_at = now()
WHERE id = $1
RETURNING id, title, target_language_id, ui_language_id, owner_id, is_published, visibility, created_at, updated_at;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;
