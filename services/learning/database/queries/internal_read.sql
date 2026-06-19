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

-- name: CountActiveEnrollments :one
SELECT COUNT(*)::int AS count FROM course_enrollments WHERE status = 'active';
