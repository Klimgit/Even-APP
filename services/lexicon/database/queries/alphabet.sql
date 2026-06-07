-- name: ListAlphabetByLanguageID :many
SELECT id, language_id, character, upper_char, sort_order, label, transcription
FROM alphabet_letters
WHERE language_id = $1
ORDER BY sort_order, character;

-- name: GetAlphabetLetter :one
SELECT id, language_id, character, upper_char, sort_order, label, transcription
FROM alphabet_letters
WHERE id = $1;

-- name: CreateAlphabetLetter :one
INSERT INTO alphabet_letters (language_id, character, upper_char, sort_order, label, transcription)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, language_id, character, upper_char, sort_order, label, transcription;

-- name: UpdateAlphabetLetter :one
UPDATE alphabet_letters
SET
    character = COALESCE(sqlc.narg('character'), character),
    upper_char = COALESCE(sqlc.narg('upper_char'), upper_char),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    label = COALESCE(sqlc.narg('label'), label),
    transcription = COALESCE(sqlc.narg('transcription'), transcription)
WHERE id = sqlc.arg('id')
RETURNING id, language_id, character, upper_char, sort_order, label, transcription;

-- name: DeleteAlphabetLetter :exec
DELETE FROM alphabet_letters WHERE id = $1;

-- name: SetAlphabetLetterSortOrder :exec
UPDATE alphabet_letters SET sort_order = $2 WHERE id = $1;
