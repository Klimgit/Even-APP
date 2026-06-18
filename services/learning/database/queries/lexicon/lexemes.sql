-- name: ListLexemesByIDs :many
SELECT id, language_id, lemma, part_of_speech, notes, created_by, created_at, updated_at
FROM lexemes
WHERE id = ANY($1::uuid[]);
