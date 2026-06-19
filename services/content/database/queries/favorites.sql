-- name: ListFavoriteBlockTypes :many
SELECT block_type FROM user_favorite_block_types WHERE user_id = $1 ORDER BY block_type;

-- name: AddFavoriteBlockType :exec
INSERT INTO user_favorite_block_types (user_id, block_type) VALUES ($1, $2)
ON CONFLICT (user_id, block_type) DO NOTHING;

-- name: RemoveFavoriteBlockType :exec
DELETE FROM user_favorite_block_types WHERE user_id = $1 AND block_type = $2;
