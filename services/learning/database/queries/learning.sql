-- name: CreateEnrollment :one
INSERT INTO course_enrollments (user_id, course_id, enrolled_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEnrollmentByUserAndCourse :one
SELECT * FROM course_enrollments
WHERE user_id = $1 AND course_id = $2;

-- name: GetEnrollmentByID :one
SELECT * FROM course_enrollments WHERE id = $1;

-- name: ListEnrollmentsByUser :many
SELECT * FROM course_enrollments
WHERE user_id = $1 AND status = 'active'
ORDER BY enrolled_at DESC;

-- name: HasEnrollment :one
SELECT EXISTS (
    SELECT 1 FROM course_enrollments
    WHERE user_id = $1 AND course_id = $2 AND status = 'active'
) AS enrolled;

-- name: UpsertPublishedLessonSnapshot :exec
INSERT INTO published_lesson_snapshots (lesson_id, course_id, version, snapshot, published_at, updated_at)
VALUES ($1, $2, $3, $4, now(), now())
ON CONFLICT (lesson_id) DO UPDATE SET
    course_id = EXCLUDED.course_id,
    version = EXCLUDED.version,
    snapshot = EXCLUDED.snapshot,
    updated_at = now();

-- name: GetPublishedLessonSnapshot :one
SELECT * FROM published_lesson_snapshots WHERE lesson_id = $1;

-- name: ListPublishedLessonSnapshotsByCourse :many
SELECT * FROM published_lesson_snapshots
WHERE course_id = $1
ORDER BY (snapshot->>'sort_order')::int;

-- name: GetUserBlockProgress :one
SELECT * FROM user_block_progress
WHERE user_id = $1 AND lesson_block_id = $2;

-- name: ListUserBlockProgressForLesson :many
SELECT ubp.*
FROM user_block_progress ubp
JOIN (
    SELECT (elem->>'id')::uuid AS block_id
    FROM published_lesson_snapshots pls,
         jsonb_array_elements(pls.snapshot->'blocks') AS elem
    WHERE pls.lesson_id = $2
) blocks ON blocks.block_id = ubp.lesson_block_id
WHERE ubp.user_id = $1;

-- name: UpsertUserBlockProgress :one
INSERT INTO user_block_progress (user_id, lesson_block_id, status, score, attempts, last_attempt_at, time_spent_seconds)
VALUES ($1, $2, $3, $4, $5, now(), $6)
ON CONFLICT (user_id, lesson_block_id) DO UPDATE SET
    status = EXCLUDED.status,
    score = EXCLUDED.score,
    attempts = EXCLUDED.attempts,
    last_attempt_at = now(),
    time_spent_seconds = user_block_progress.time_spent_seconds + EXCLUDED.time_spent_seconds
RETURNING *;

-- name: InsertBlockAttempt :exec
INSERT INTO block_attempts (user_id, lesson_block_id, sub_item_index, is_correct, response, context)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpsertReviewItemOnFailure :exec
INSERT INTO user_review_items (
    user_id, lesson_block_id, sub_item_index, status, failure_count,
    consecutive_ok, first_failed_at, last_failed_at, last_attempt_at, due_at,
    source_lesson_id, course_id
) VALUES (
    $1, $2, $3, 'pending', 1, 0, now(), now(), now(), $4, $5, $6
)
ON CONFLICT (user_id, lesson_block_id, sub_item_index) DO UPDATE SET
    status = 'pending',
    failure_count = user_review_items.failure_count + 1,
    consecutive_ok = 0,
    last_failed_at = now(),
    last_attempt_at = now(),
    due_at = $4;

-- name: UpdateReviewItemOnSuccess :exec
UPDATE user_review_items SET
    consecutive_ok = consecutive_ok + 1,
    last_attempt_at = now(),
    status = CASE WHEN consecutive_ok + 1 >= 2 THEN 'mastered' ELSE status END
WHERE user_id = $1 AND lesson_block_id = $2 AND sub_item_index = $3;

-- name: ListReviewItems :many
SELECT * FROM user_review_items
WHERE user_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
  AND (NOT sqlc.arg('due_only')::bool OR (status = 'pending' AND due_at <= now()))
ORDER BY due_at ASC;

-- name: CountReviewItems :one
SELECT
    COUNT(*) FILTER (WHERE status = 'pending')::int AS pending_count,
    COUNT(*) FILTER (WHERE status = 'pending' AND due_at <= now())::int AS due_count
FROM user_review_items
WHERE user_id = $1;

-- name: GetNextDueReviewItem :one
SELECT * FROM user_review_items
WHERE user_id = $1 AND status = 'pending' AND due_at <= now()
ORDER BY due_at ASC
LIMIT 1;

-- name: UpsertVocabularyOnSuccess :exec
INSERT INTO user_vocabulary (user_id, lexeme_id, course_id, first_seen_at, mastery)
VALUES ($1, $2, $3, now(), $4)
ON CONFLICT (user_id, lexeme_id) DO UPDATE SET
    mastery = GREATEST(user_vocabulary.mastery, EXCLUDED.mastery);

-- name: ListVocabulary :many
SELECT * FROM user_vocabulary
WHERE user_id = $1
  AND (sqlc.narg('course_id')::uuid IS NULL OR course_id = sqlc.narg('course_id')::uuid)
ORDER BY first_seen_at DESC;

-- name: CountCompletedGradableBlocksForCourse :one
SELECT
    COUNT(*) FILTER (WHERE ubp.status = 'completed')::int AS completed,
    COUNT(*)::int AS total
FROM user_block_progress ubp
JOIN course_enrollments ce ON ce.user_id = ubp.user_id AND ce.course_id = $2
WHERE ubp.user_id = $1
  AND ubp.lesson_block_id IN (
    SELECT (elem->>'id')::uuid
    FROM published_lesson_snapshots pls,
         jsonb_array_elements(pls.snapshot->'blocks') AS elem
    WHERE pls.course_id = $2
      AND COALESCE((elem->>'is_gradable')::bool, false) = true
  );

-- name: CountCompletedGradableBlocksForLesson :one
SELECT
    COUNT(*) FILTER (WHERE ubp.status = 'completed')::int AS completed,
    COUNT(*)::int AS total
FROM user_block_progress ubp
WHERE ubp.user_id = $1
  AND ubp.lesson_block_id IN (
    SELECT (elem->>'id')::uuid
    FROM published_lesson_snapshots pls,
         jsonb_array_elements(pls.snapshot->'blocks') AS elem
    WHERE pls.lesson_id = $2
      AND COALESCE((elem->>'is_gradable')::bool, false) = true
  );

-- name: GetLessonTitleFromSnapshot :one
SELECT (snapshot->>'title')::text AS title
FROM published_lesson_snapshots
WHERE lesson_id = $1;

-- name: GetProgressSummary :one
SELECT
    (SELECT COUNT(*)::int FROM course_enrollments ce WHERE ce.user_id = $1 AND ce.status = 'active') AS enrolled_courses,
    (SELECT COUNT(*)::int FROM user_vocabulary uv WHERE uv.user_id = $1) AS dictionary_words,
    (SELECT COUNT(*)::int FROM user_review_items uri WHERE uri.user_id = $1 AND uri.status = 'pending' AND uri.due_at <= now()) AS review_due,
    (SELECT COUNT(*)::int FROM user_block_progress ubp WHERE ubp.user_id = $1 AND ubp.status = 'completed') AS completed_blocks,
    (SELECT COUNT(*)::int
     FROM published_lesson_snapshots pls
     JOIN course_enrollments ce ON ce.course_id = pls.course_id AND ce.user_id = $1 AND ce.status = 'active'
     WHERE (
         SELECT COUNT(*)::int
         FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
         WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
     ) > 0
     AND (
         SELECT COUNT(*) FILTER (WHERE ubp.status = 'completed')::int
         FROM user_block_progress ubp
         WHERE ubp.user_id = $1
           AND ubp.lesson_block_id IN (
               SELECT (elem->>'id')::uuid
               FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
               WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
           )
     ) = (
         SELECT COUNT(*)::int
         FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
         WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
     )) AS completed_lessons,
    (SELECT COUNT(*)::int
     FROM published_lesson_snapshots pls
     JOIN course_enrollments ce ON ce.course_id = pls.course_id AND ce.user_id = $1 AND ce.status = 'active'
     WHERE (
         SELECT COUNT(*)::int
         FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
         WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
     ) > 0
     AND (
         SELECT COUNT(*) FILTER (WHERE ubp.status = 'completed')::int
         FROM user_block_progress ubp
         WHERE ubp.user_id = $1
           AND ubp.lesson_block_id IN (
               SELECT (elem->>'id')::uuid
               FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
               WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
           )
     ) > 0
     AND (
         SELECT COUNT(*) FILTER (WHERE ubp.status = 'completed')::int
         FROM user_block_progress ubp
         WHERE ubp.user_id = $1
           AND ubp.lesson_block_id IN (
               SELECT (elem->>'id')::uuid
               FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
               WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
           )
     ) < (
         SELECT COUNT(*)::int
         FROM jsonb_array_elements(pls.snapshot->'blocks') AS elem
         WHERE COALESCE((elem->>'is_gradable')::bool, false) = true
     )) AS in_progress_lessons,
    (SELECT COALESCE(AVG(ubp.score), 0)::float8
     FROM user_block_progress ubp
     WHERE ubp.user_id = $1 AND ubp.status = 'completed') AS average_score,
    (SELECT COALESCE(SUM(ubp.time_spent_seconds), 0)::int
     FROM user_block_progress ubp
     WHERE ubp.user_id = $1) AS time_spent_seconds;

-- name: ListEnrollmentsByCourse :many
SELECT user_id, course_id, status, enrolled_at, enrolled_by
FROM course_enrollments
WHERE course_id = $1 AND status = 'active'
ORDER BY enrolled_at DESC;
