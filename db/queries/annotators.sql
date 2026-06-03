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