-- name: ListDemoNotes :many
SELECT id, text, created_at
FROM demo_notes
ORDER BY created_at ASC;
