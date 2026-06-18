-- name: ListSectionsByLessonID :many
SELECT id, lesson_id, title, sort_order, section_kind
FROM lesson_sections
WHERE lesson_id = $1
ORDER BY sort_order, title;

-- name: GetSectionByID :one
SELECT id, lesson_id, title, sort_order, section_kind
FROM lesson_sections
WHERE id = $1;

-- name: GetSectionCourseOwner :one
SELECT c.owner_id
FROM lesson_sections s
JOIN lessons l ON l.id = s.lesson_id
JOIN courses c ON c.id = l.course_id
WHERE s.id = $1;

-- name: CreateSection :one
INSERT INTO lesson_sections (lesson_id, title, sort_order, section_kind)
VALUES ($1, $2, $3, $4)
RETURNING id, lesson_id, title, sort_order, section_kind;

-- name: UpdateSection :one
UPDATE lesson_sections
SET
    title = COALESCE(sqlc.narg('title'), title),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    section_kind = COALESCE(sqlc.narg('section_kind'), section_kind)
WHERE id = sqlc.arg('id')
RETURNING id, lesson_id, title, sort_order, section_kind;

-- name: DeleteSection :exec
DELETE FROM lesson_sections WHERE id = $1;

-- name: SetSectionSortOrder :exec
UPDATE lesson_sections SET sort_order = $2 WHERE id = $1;

-- name: MaxSectionSortOrder :one
SELECT COALESCE(MAX(sort_order), -1)::int AS max_sort
FROM lesson_sections
WHERE lesson_id = $1;
