-- name: CountLexemesByLanguage :one
SELECT count(*)::int AS count
FROM lexemes
WHERE language_id = $1
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR lemma ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: ListLexemesByLanguage :many
SELECT id, language_id, lemma, part_of_speech, notes, created_by, created_at, updated_at
FROM lexemes
WHERE language_id = $1
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR lemma ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY lemma
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetLexeme :one
SELECT id, language_id, lemma, part_of_speech, notes, created_by, created_at, updated_at
FROM lexemes
WHERE id = $1;

-- name: CreateLexeme :one
INSERT INTO lexemes (language_id, lemma, part_of_speech, notes, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, language_id, lemma, part_of_speech, notes, created_by, created_at, updated_at;

-- name: UpdateLexeme :one
UPDATE lexemes
SET
    lemma = COALESCE(sqlc.narg('lemma'), lemma),
    part_of_speech = sqlc.narg('part_of_speech'),
    notes = sqlc.narg('notes'),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, language_id, lemma, part_of_speech, notes, created_by, created_at, updated_at;

-- name: DeleteLexeme :exec
DELETE FROM lexemes WHERE id = $1;

-- name: ListLexemeForms :many
SELECT id, lexeme_id, form, tags
FROM lexeme_forms
WHERE lexeme_id = $1
ORDER BY form;

-- name: GetLexemeForm :one
SELECT id, lexeme_id, form, tags
FROM lexeme_forms
WHERE id = $1;

-- name: CreateLexemeForm :one
INSERT INTO lexeme_forms (lexeme_id, form, tags)
VALUES ($1, $2, $3)
RETURNING id, lexeme_id, form, tags;

-- name: UpdateLexemeForm :one
UPDATE lexeme_forms
SET
    form = COALESCE(sqlc.narg('form'), form),
    tags = COALESCE(sqlc.narg('tags'), tags)
WHERE id = sqlc.arg('id')
RETURNING id, lexeme_id, form, tags;

-- name: DeleteLexemeForm :exec
DELETE FROM lexeme_forms WHERE id = $1;

-- name: ListLexemeTranslations :many
SELECT id, source_lexeme_id, target_language_id, text, target_lexeme_id
FROM lexeme_translations
WHERE source_lexeme_id = $1;

-- name: GetLexemeTranslation :one
SELECT id, source_lexeme_id, target_language_id, text, target_lexeme_id
FROM lexeme_translations
WHERE id = $1;

-- name: CreateLexemeTranslation :one
INSERT INTO lexeme_translations (source_lexeme_id, target_language_id, text, target_lexeme_id)
VALUES ($1, $2, $3, $4)
RETURNING id, source_lexeme_id, target_language_id, text, target_lexeme_id;

-- name: DeleteLexemeTranslation :exec
DELETE FROM lexeme_translations WHERE id = $1;

-- name: ListLexemeMedia :many
SELECT id, lexeme_id, form_id, media_asset_id, kind, label, is_primary, sort_order
FROM lexeme_media
WHERE lexeme_id = $1
ORDER BY sort_order, id;

-- name: GetLexemeMedia :one
SELECT id, lexeme_id, form_id, media_asset_id, kind, label, is_primary, sort_order
FROM lexeme_media
WHERE id = $1;

-- name: CreateLexemeMedia :one
INSERT INTO lexeme_media (lexeme_id, form_id, media_asset_id, kind, label, is_primary, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, lexeme_id, form_id, media_asset_id, kind, label, is_primary, sort_order;

-- name: DeleteLexemeMedia :exec
DELETE FROM lexeme_media WHERE id = $1;

-- name: ClearPrimaryLexemeMedia :exec
UPDATE lexeme_media SET is_primary = false WHERE lexeme_id = $1 AND kind = $2;
