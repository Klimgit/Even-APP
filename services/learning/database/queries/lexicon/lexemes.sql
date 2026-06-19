-- name: ListLexemesByIDs :many
SELECT id, language_id, lemma, part_of_speech, notes, created_by, created_at, updated_at
FROM lexemes
WHERE id = ANY($1::uuid[]);

-- name: FilterLexemeIDsBySearch :many
SELECT DISTINCT l.id
FROM lexemes l
LEFT JOIN lexeme_translations t ON t.source_lexeme_id = l.id
WHERE l.id = ANY($1::uuid[])
  AND (
    l.lemma ILIKE '%' || $2 || '%'
    OR t.text ILIKE '%' || $2 || '%'
  );
