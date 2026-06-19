-- name: ListEnrollmentsByCourse :many
SELECT user_id, course_id, status, enrolled_at, enrolled_by
FROM course_enrollments
WHERE course_id = $1 AND status = 'active'
ORDER BY enrolled_at DESC;

-- name: GetStudentLessonProgress :many
SELECT
    pls.lesson_id,
    (pls.snapshot->>'title')::text AS title,
    COUNT(blocks.block_id)::int AS total_blocks,
    COUNT(*) FILTER (WHERE ubp.status = 'completed')::int AS completed_blocks,
    COALESCE(AVG(ubp.score) FILTER (WHERE ubp.status = 'completed'), 0)::real AS score_avg
FROM published_lesson_snapshots pls
CROSS JOIN LATERAL (
    SELECT (elem->>'id')::uuid AS block_id
    FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
) blocks
LEFT JOIN user_block_progress ubp
    ON ubp.user_id = $1 AND ubp.lesson_block_id = blocks.block_id
WHERE pls.course_id = $2
GROUP BY pls.lesson_id, pls.snapshot->>'title', (pls.snapshot->>'sort_order')::int
ORDER BY (pls.snapshot->>'sort_order')::int;

-- name: CreateEnrollment :one
INSERT INTO course_enrollments (user_id, course_id, enrolled_by)
VALUES ($1, $2, $3)
RETURNING id, user_id, course_id, status, enrolled_at, enrolled_by;

-- name: GetEnrollmentByUserAndCourse :one
SELECT id, user_id, course_id, status, enrolled_at, enrolled_by
FROM course_enrollments
WHERE user_id = $1 AND course_id = $2;

-- name: UpsertPublishedLessonSnapshot :exec
INSERT INTO published_lesson_snapshots (lesson_id, course_id, version, snapshot, published_at, updated_at)
VALUES ($1, $2, $3, $4, now(), now())
ON CONFLICT (lesson_id) DO UPDATE SET
    course_id = EXCLUDED.course_id,
    version = EXCLUDED.version,
    snapshot = EXCLUDED.snapshot,
    updated_at = now();
