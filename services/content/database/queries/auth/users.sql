-- name: ListUsersByIDs :many
SELECT id, email, display_name, role, is_admin, created_at
FROM users
WHERE id = ANY($1::uuid[]);

-- name: GetUserByEmail :one
SELECT id, email, display_name, role, is_admin, created_at
FROM users
WHERE email = $1;
