-- name: GetLanguageIDByCode :one
SELECT id
FROM languages
WHERE code = $1;

-- name: LanguageExists :one
SELECT EXISTS(SELECT 1 FROM languages WHERE id = $1)::bool AS exists;
