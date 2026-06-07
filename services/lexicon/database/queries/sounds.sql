-- name: ListSoundsByLanguageID :many
SELECT id, language_id, ipa, description, audio_key
FROM sounds
WHERE language_id = $1
ORDER BY ipa NULLS LAST, id;

-- name: GetSound :one
SELECT id, language_id, ipa, description, audio_key
FROM sounds
WHERE id = $1;

-- name: CreateSound :one
INSERT INTO sounds (language_id, ipa, description, audio_key)
VALUES ($1, $2, $3, $4)
RETURNING id, language_id, ipa, description, audio_key;

-- name: UpdateSound :one
UPDATE sounds
SET
    ipa = sqlc.narg('ipa'),
    description = sqlc.narg('description'),
    audio_key = sqlc.narg('audio_key')
WHERE id = sqlc.arg('id')
RETURNING id, language_id, ipa, description, audio_key;

-- name: DeleteSound :exec
DELETE FROM sounds WHERE id = $1;

-- name: LinkLetterSound :exec
INSERT INTO letter_sounds (letter_id, sound_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnlinkLetterSound :exec
DELETE FROM letter_sounds WHERE letter_id = $1 AND sound_id = $2;
