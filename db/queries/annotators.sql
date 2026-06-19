-- name: CreateAnnotator :one
INSERT INTO annotators (kind, model_id, label)
VALUES ($1, $2, $3)
ON CONFLICT (label) DO UPDATE SET label = EXCLUDED.label
RETURNING id;

-- name: GetAnnotatorByLabel :one
SELECT * FROM annotators WHERE label = $1;

-- name: GetLLMAnnotatorByModel :one
SELECT * FROM annotators WHERE kind = 'llm' and model_id = $1;

-- name: ListAnnotators :many
SELECT * FROM annotators ORDER BY kind, label;

-- name: ListAnnotatorsByIDs :many
SELECT * FROM annotators WHERE id = ANY(sqlc.arg(ids)::int[]) ORDER BY id;

-- name: ListLLMAnnotators :many
-- the active LLM panel fetch from the DB with each annotator with its model slug
SELECT
    a.id,
    a.label,
    m.slug,
    m.family
FROM annotators a
JOIN models m ON m.id = a.model_id
WHERE a.kind = 'llm' AND m.active
ORDER BY a.id;