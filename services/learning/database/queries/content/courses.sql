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
    ic.code AS invite_code,
    c.created_at,
    c.updated_at
FROM courses c
JOIN course_invite_codes ic ON ic.course_id = c.id
WHERE ic.code = $1 AND c.is_published = true;

-- name: GetCourseByID :one
SELECT
    c.id,
    c.title,
    c.target_language_id,
    ''::text AS target_language_code,
    ''::text AS target_language_name,
    c.ui_language_id,
    c.owner_id,
    c.is_published,
    ic.code AS invite_code,
    c.created_at,
    c.updated_at
FROM courses c
LEFT JOIN course_invite_codes ic ON ic.course_id = c.id
WHERE c.id = $1;

-- name: ListPublishedLessonsByCourse :many
SELECT id, course_id, title, sort_order, version, status, published_at, updated_at
FROM lessons
WHERE course_id = $1 AND status = 'published'
ORDER BY sort_order ASC;

-- name: GetPublishedLessonByID :one
SELECT id, course_id, title, sort_order, version, status, published_at, updated_at
FROM lessons
WHERE id = $1 AND status = 'published';

-- name: ListLessonSections :many
SELECT id, lesson_id, title, sort_order, section_kind
FROM lesson_sections
WHERE lesson_id = $1
ORDER BY sort_order ASC;

-- name: ListLessonBlocks :many
SELECT id, lesson_id, section_id, sort_order, display_label, title, block_type, config, is_homework
FROM lesson_blocks
WHERE lesson_id = $1
ORDER BY sort_order ASC;

-- name: GetLessonBlock :one
SELECT lb.id, lb.lesson_id, lb.section_id, lb.sort_order, lb.display_label, lb.title, lb.block_type, lb.config, lb.is_homework,
       l.course_id, l.id AS lesson_id, l.title AS lesson_title, l.status AS lesson_status
FROM lesson_blocks lb
JOIN lessons l ON l.id = lb.lesson_id
WHERE lb.id = $1;

-- name: ListBlockLexemeRefs :many
SELECT lesson_block_id, lexeme_id, form_id, role
FROM block_lexeme_refs
WHERE lesson_block_id = $1;

-- name: GetLessonTitle :one
SELECT title FROM lessons WHERE id = $1;
