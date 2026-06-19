-- name: ListBlocksByCourseID :many
SELECT b.id, b.lesson_id, b.section_id, b.sort_order, b.display_label, b.title, b.block_type, b.config, b.is_homework,
       l.title AS lesson_title, l.sort_order AS lesson_sort_order
FROM lesson_blocks b
JOIN lessons l ON l.id = b.lesson_id
WHERE l.course_id = $1
ORDER BY l.sort_order, b.sort_order, b.id;
