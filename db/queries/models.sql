-- name: UpsertModel :one
INSERT INTO models (family, slug, name, modalities)
VALUES ($1, $2, $3, $4)
ON CONFLICT (slug) DO UPDATE
    SET family     = EXCLUDED.family,
        name       = EXCLUDED.name,
        modalities = EXCLUDED.modalities
RETURNING id;

-- name: ListActiveModels :many
SELECT * FROM models where active ORDER BY family, slug;

-- name: ListActiveModelsByModality :many
SELECT *
FROM models
WHERE active
    AND $1::varchar = ANY(modalities)
ORDER BY family;

-- name: GetModelBySlug :one
SELECT * FROM models where slug = $1;