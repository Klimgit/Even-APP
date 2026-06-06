-- name: SaveRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token_hash = $1;

-- name: GetUserIDByRefreshHash :one
SELECT user_id
FROM refresh_tokens
WHERE token_hash = $1
  AND expires_at > now();
