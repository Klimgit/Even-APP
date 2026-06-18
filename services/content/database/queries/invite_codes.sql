-- name: GetInviteCodeByCourseID :one
SELECT course_id, code, created_at
FROM course_invite_codes
WHERE course_id = $1;

-- name: UpsertInviteCode :one
INSERT INTO course_invite_codes (course_id, code)
VALUES ($1, $2)
ON CONFLICT (course_id) DO UPDATE SET code = EXCLUDED.code, created_at = now()
RETURNING course_id, code, created_at;
