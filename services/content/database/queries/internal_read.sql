-- Internal read queries for service-to-service HTTP handlers.

-- name: GetCourseByInviteCode :one
SELECT
    c.id,
    c.title,
    c.target_language_id,
    ''::text AS target_language_code,
    ''::text AS target_language_name,
    c.ui_language_id,
    c.owner_id,
    c.is_published,
    c.visibility,
    ic.code AS invite_code
FROM courses c
JOIN course_invite_codes ic ON ic.course_id = c.id
WHERE ic.code = $1 AND c.is_published = true;

-- name: ListPublishedCourses :many
SELECT
    c.id,
    c.title,
    c.target_language_id,
    ''::text AS target_language_code,
    ''::text AS target_language_name,
    c.ui_language_id,
    c.owner_id,
    c.is_published,
    c.visibility
FROM courses c
WHERE c.is_published = true AND c.visibility = 'public'
ORDER BY c.updated_at DESC;

-- name: CountPublishedCourses :one
SELECT COUNT(*)::int AS count FROM courses WHERE is_published = true;

-- name: GetLessonTitle :one
SELECT title FROM lessons WHERE id = $1;

-- name: GetLessonBlockWithCourse :one
SELECT lb.id, lb.lesson_id, lb.section_id, lb.sort_order, lb.display_label, lb.title, lb.block_type, lb.config, lb.is_homework,
       l.course_id, l.title AS lesson_title
FROM lesson_blocks lb
JOIN lessons l ON l.id = lb.lesson_id
WHERE lb.id = $1;

-- name: ListPublishedLessonsByCourse :many
SELECT id, course_id, title, sort_order, version, status
FROM lessons
WHERE course_id = $1 AND status = 'published'
ORDER BY sort_order ASC;
