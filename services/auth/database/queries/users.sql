-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, role)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, display_name, role, is_admin, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, display_name, role, is_admin, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, display_name, role, is_admin, created_at
FROM users
WHERE id = $1;

-- name: CountUsers :one
SELECT count(*)::int AS count FROM users;

-- name: UserStats :one
SELECT
  count(*)::int AS total_users,
  count(*) FILTER (WHERE role = 'student')::int AS students,
  count(*) FILTER (WHERE role = 'teacher')::int AS teachers,
  count(*) FILTER (WHERE is_admin = true)::int AS admins
FROM users;
