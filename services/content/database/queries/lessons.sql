-- name: ListLessonsByCourseID :many
SELECT id, course_id, title, sort_order, version, status, published_at, updated_at
FROM lessons
WHERE course_id = $1
ORDER BY sort_order, title;

-- name: GetLessonByID :one
SELECT id, course_id, title, sort_order, version, status, published_at, updated_at
FROM lessons
WHERE id = $1;

-- name: GetLessonCourseOwner :one
SELECT c.owner_id
FROM lessons l
JOIN courses c ON c.id = l.course_id
WHERE l.id = $1;

-- name: CreateLesson :one
INSERT INTO lessons (course_id, title, sort_order)
VALUES ($1, $2, $3)
RETURNING id, course_id, title, sort_order, version, status, published_at, updated_at;

-- name: UpdateLesson :one
UPDATE lessons
SET
    title = COALESCE(sqlc.narg('title'), title),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    version = version + 1,
    updated_at = now()
WHERE id = sqlc.arg('id') AND version = sqlc.arg('expected_version')
RETURNING id, course_id, title, sort_order, version, status, published_at, updated_at;

-- name: PublishLesson :one
UPDATE lessons
SET status = 'published', published_at = now(), updated_at = now()
WHERE id = $1
RETURNING id, course_id, title, sort_order, version, status, published_at, updated_at;

-- name: DeleteLesson :exec
DELETE FROM lessons WHERE id = $1;

-- name: MaxLessonSortOrder :one
SELECT COALESCE(MAX(sort_order), -1)::int AS max_sort
FROM lessons
WHERE course_id = $1;
