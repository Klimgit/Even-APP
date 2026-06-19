-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, role)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, display_name, role, is_admin, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, display_name, role, is_admin, created_at
FROM users
WHERE email = $1;

-- name: ListUsersByIDs :many
SELECT id, email, display_name, role, is_admin, created_at
FROM users
WHERE id = ANY($1::uuid[]);

-- name: GetUserByID :one
SELECT id, email, password_hash, display_name, role, is_admin, created_at
FROM users
WHERE id = $1;

-- name: CountUsers :one
SELECT count(*)::int AS count FROM users;

-- name: ListUsers :many
SELECT id, email, password_hash, display_name, role, is_admin, created_at
FROM users
WHERE ($1::text IS NULL OR $1 = '' OR email ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR $2 = '' OR role = $2)
ORDER BY created_at DESC
LIMIT $4 OFFSET $3;

-- name: CountUsersFiltered :one
SELECT count(*)::int AS count
FROM users
WHERE ($1::text IS NULL OR $1 = '' OR email ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR $2 = '' OR role = $2);

-- name: UpdateUserPlatform :one
UPDATE users
SET
  role = COALESCE(sqlc.narg('role'), role),
  is_admin = COALESCE(sqlc.narg('is_admin'), is_admin)
WHERE id = $1
RETURNING id, email, password_hash, display_name, role, is_admin, created_at;

-- name: UserStats :one
SELECT
  count(*)::int AS total_users,
  count(*) FILTER (WHERE role = 'student')::int AS students,
  count(*) FILTER (WHERE role = 'teacher')::int AS teachers,
  count(*) FILTER (WHERE is_admin = true)::int AS admins
FROM users;
