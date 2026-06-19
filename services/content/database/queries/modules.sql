-- name: ListModulesByCourseID :many
SELECT id, course_id, title, sort_order, created_at, updated_at
FROM course_modules
WHERE course_id = $1
ORDER BY sort_order, title;

-- name: GetModuleByID :one
SELECT id, course_id, title, sort_order, created_at, updated_at
FROM course_modules
WHERE id = $1;

-- name: GetModuleCourseOwner :one
SELECT c.owner_id
FROM course_modules m
JOIN courses c ON c.id = m.course_id
WHERE m.id = $1;

-- name: CreateModule :one
INSERT INTO course_modules (course_id, title, sort_order)
VALUES ($1, $2, $3)
RETURNING id, course_id, title, sort_order, created_at, updated_at;

-- name: UpdateModule :one
UPDATE course_modules
SET
    title = COALESCE(sqlc.narg('title'), title),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, course_id, title, sort_order, created_at, updated_at;

-- name: DeleteModule :exec
DELETE FROM course_modules WHERE id = $1;

-- name: MaxModuleSortOrder :one
SELECT COALESCE(MAX(sort_order), -1)::int AS max_sort
FROM course_modules
WHERE course_id = $1;

-- name: SetModuleSortOrder :exec
UPDATE course_modules SET sort_order = $2, updated_at = now() WHERE id = $1;

-- name: CountLessonsInModule :one
SELECT COUNT(*)::int AS count FROM lessons WHERE module_id = $1;
