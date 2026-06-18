-- name: ListBlocksByLessonID :many
SELECT id, lesson_id, section_id, sort_order, display_label, title, block_type, config, is_homework
FROM lesson_blocks
WHERE lesson_id = $1
ORDER BY sort_order, id;

-- name: ListBlocksByCourseID :many
SELECT b.id, b.lesson_id, b.section_id, b.sort_order, b.display_label, b.title, b.block_type, b.config, b.is_homework,
       l.title AS lesson_title, l.sort_order AS lesson_sort_order
FROM lesson_blocks b
JOIN lessons l ON l.id = b.lesson_id
WHERE l.course_id = $1
ORDER BY l.sort_order, b.sort_order, b.id;

-- name: GetBlockByID :one
SELECT id, lesson_id, section_id, sort_order, display_label, title, block_type, config, is_homework
FROM lesson_blocks
WHERE id = $1;

-- name: GetBlockCourseOwner :one
SELECT c.owner_id
FROM lesson_blocks b
JOIN lessons l ON l.id = b.lesson_id
JOIN courses c ON c.id = l.course_id
WHERE b.id = $1;

-- name: CreateBlock :one
INSERT INTO lesson_blocks (lesson_id, section_id, sort_order, display_label, title, block_type, config, is_homework)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, lesson_id, section_id, sort_order, display_label, title, block_type, config, is_homework;

-- name: UpdateBlock :one
UPDATE lesson_blocks
SET
    section_id = COALESCE(sqlc.narg('section_id'), section_id),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    display_label = COALESCE(sqlc.narg('display_label'), display_label),
    title = COALESCE(sqlc.narg('title'), title),
    block_type = COALESCE(sqlc.narg('block_type'), block_type),
    config = COALESCE(sqlc.narg('config'), config),
    is_homework = COALESCE(sqlc.narg('is_homework'), is_homework)
WHERE id = sqlc.arg('id')
RETURNING id, lesson_id, section_id, sort_order, display_label, title, block_type, config, is_homework;

-- name: DeleteBlock :exec
DELETE FROM lesson_blocks WHERE id = $1;

-- name: SetBlockSortOrder :exec
UPDATE lesson_blocks SET sort_order = $2 WHERE id = $1;

-- name: MaxBlockSortOrder :one
SELECT COALESCE(MAX(sort_order), -1)::int AS max_sort
FROM lesson_blocks
WHERE lesson_id = $1;
