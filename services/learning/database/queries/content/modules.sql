-- name: ListModulesByCourseID :many
SELECT id, course_id, title, sort_order, created_at, updated_at
FROM course_modules
WHERE course_id = $1
ORDER BY sort_order, title;

-- name: GetModuleByID :one
SELECT id, course_id, title, sort_order, created_at, updated_at
FROM course_modules
WHERE id = $1;

-- name: ListPublishedLessonsByModule :many
SELECT id, course_id, module_id, title, sort_order, version, status, published_at, updated_at
FROM lessons
WHERE module_id = $1 AND status = 'published'
ORDER BY sort_order ASC;
