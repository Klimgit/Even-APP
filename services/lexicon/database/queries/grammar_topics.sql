-- name: ListGrammarTopicsByLanguageCode :many
SELECT gt.id, gt.language_id, gt.title, gt.description, gt.sort_order, gt.created_at, gt.updated_at
FROM grammar_topics gt
JOIN languages l ON l.id = gt.language_id
WHERE l.code = $1
ORDER BY gt.sort_order, gt.title;

-- name: CreateGrammarTopic :one
INSERT INTO grammar_topics (language_id, title, description, sort_order)
VALUES ($1, $2, $3, COALESCE(sqlc.narg('sort_order'), 0))
RETURNING id, language_id, title, description, sort_order, created_at, updated_at;

-- name: PatchGrammarTopic :one
UPDATE grammar_topics
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, language_id, title, description, sort_order, created_at, updated_at;

-- name: DeleteGrammarTopic :exec
DELETE FROM grammar_topics WHERE id = $1;
