-- name: GetLanguageIDByCode :one
SELECT id
FROM languages
WHERE code = $1;
