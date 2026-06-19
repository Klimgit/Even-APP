-- name: GetLanguageByID :one
SELECT id, code, name, native_name, direction, is_active
FROM languages
WHERE id = $1;

-- name: ListLanguagesByIDs :many
SELECT id, code, name, native_name, direction, is_active
FROM languages
WHERE id = ANY($1::uuid[]);

-- name: ListLexemesByIDs :many
SELECT id, lemma, part_of_speech
FROM lexemes
WHERE id = ANY($1::uuid[]);
