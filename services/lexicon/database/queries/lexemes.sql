-- name: CountLexemesByLanguage :one
SELECT count(*)::int AS count
FROM lexemes
WHERE language_id = $1
  AND scope = sqlc.arg('scope')
  AND (
    sqlc.narg('owner_id')::uuid IS NULL
    OR owner_id = sqlc.narg('owner_id')
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR lemma ILIKE '%' || sqlc.narg('search') || '%'
    OR EXISTS (
      SELECT 1 FROM lexeme_translations t
      WHERE t.source_lexeme_id = lexemes.id
        AND t.text ILIKE '%' || sqlc.narg('search') || '%'
    )
  );

-- name: ListLexemesByLanguage :many
SELECT id, language_id, lemma, part_of_speech, notes, scope, owner_id, created_by, created_at, updated_at
FROM lexemes
WHERE language_id = $1
  AND scope = sqlc.arg('scope')
  AND (
    sqlc.narg('owner_id')::uuid IS NULL
    OR owner_id = sqlc.narg('owner_id')
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR lemma ILIKE '%' || sqlc.narg('search') || '%'
    OR EXISTS (
      SELECT 1 FROM lexeme_translations t
      WHERE t.source_lexeme_id = lexemes.id
        AND t.text ILIKE '%' || sqlc.narg('search') || '%'
    )
  )
ORDER BY lemma
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPickerLexemes :one
SELECT count(*)::int AS count
FROM lexemes
WHERE language_id = $1
  AND (
    scope = 'platform'
    OR (scope = 'teacher' AND owner_id = sqlc.arg('owner_id'))
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR lemma ILIKE '%' || sqlc.narg('search') || '%'
    OR EXISTS (
      SELECT 1 FROM lexeme_translations t
      WHERE t.source_lexeme_id = lexemes.id
        AND t.text ILIKE '%' || sqlc.narg('search') || '%'
    )
  );

-- name: ListPickerLexemes :many
SELECT id, language_id, lemma, part_of_speech, notes, scope, owner_id, created_by, created_at, updated_at
FROM lexemes
WHERE language_id = $1
  AND (
    scope = 'platform'
    OR (scope = 'teacher' AND owner_id = sqlc.arg('owner_id'))
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR sqlc.narg('search')::text = ''
    OR lemma ILIKE '%' || sqlc.narg('search') || '%'
    OR EXISTS (
      SELECT 1 FROM lexeme_translations t
      WHERE t.source_lexeme_id = lexemes.id
        AND t.text ILIKE '%' || sqlc.narg('search') || '%'
    )
  )
ORDER BY lemma
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetLexeme :one
SELECT id, language_id, lemma, part_of_speech, notes, scope, owner_id, created_by, created_at, updated_at
FROM lexemes
WHERE id = $1;

-- name: CreateLexeme :one
INSERT INTO lexemes (language_id, lemma, part_of_speech, notes, scope, owner_id, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, language_id, lemma, part_of_speech, notes, scope, owner_id, created_by, created_at, updated_at;

-- name: UpdateLexeme :one
UPDATE lexemes
SET
    lemma = COALESCE(sqlc.narg('lemma'), lemma),
    part_of_speech = sqlc.narg('part_of_speech'),
    notes = sqlc.narg('notes'),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, language_id, lemma, part_of_speech, notes, scope, owner_id, created_by, created_at, updated_at;

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

-- name: UpdateLexemeTranslation :one
UPDATE lexeme_translations
SET
    text = COALESCE(sqlc.narg('text'), text),
    target_language_id = COALESCE(sqlc.narg('target_language_id'), target_language_id),
    target_lexeme_id = sqlc.narg('target_lexeme_id')
WHERE id = sqlc.arg('id')
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

-- name: UpdateLexemeMedia :one
UPDATE lexeme_media
SET
    kind = COALESCE(sqlc.narg('kind'), kind),
    label = sqlc.narg('label'),
    is_primary = COALESCE(sqlc.narg('is_primary'), is_primary),
    media_asset_id = COALESCE(sqlc.narg('media_asset_id'), media_asset_id),
    form_id = sqlc.narg('form_id')
WHERE id = sqlc.arg('id')
RETURNING id, lexeme_id, form_id, media_asset_id, kind, label, is_primary, sort_order;

-- name: DeleteLexemeMedia :exec
DELETE FROM lexeme_media WHERE id = $1;

-- name: ClearPrimaryLexemeMedia :exec
UPDATE lexeme_media SET is_primary = false WHERE lexeme_id = $1 AND kind = $2;

-- name: CountLexemes :one
SELECT COUNT(*)::int AS count FROM lexemes WHERE scope = 'platform';

-- name: FilterLexemeIDsBySearch :many
SELECT DISTINCT l.id
FROM lexemes l
LEFT JOIN lexeme_translations t ON t.source_lexeme_id = l.id
WHERE l.id = ANY($1::uuid[])
  AND (
    l.lemma ILIKE '%' || $2 || '%'
    OR t.text ILIKE '%' || $2 || '%'
  );
