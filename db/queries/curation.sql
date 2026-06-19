-- name: AddIndividual :exec
INSERT INTO individuals (population_id, artifact_id)
VALUES ($1, $2)
ON CONFLICT (population_id, artifact_id) DO NOTHING;

-- name: CountIndividuals :one
SELECT count(*) FROM individuals WHERE population_id = $1;

-- name: CreatePopulation :one
INSERT INTO populations (modality, description, criteria, artifacts_cutoff)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: SelectTextReviewsFromPopulation :many
-- representative selection with common filters and the cutoff
-- criteria need dynamic SQL
SELECT
FROM artifacts a
JOIN text_review_details td ON td.artifact_id = a.id
WHERE a.modality = 'text'
    AND a.scraped_at <= $1
    AND (td.source_meta->>'hours_played')::int >= $2;

-- name: FreezeStratifiedPopulation :execrows
INSERT INTO individuals (population_id, artifact_id)
SELECT @population_id::int, ranked.artifact_id
FROM (
    SELECT a.id as artifact_id,
           row_number() OVER (
            PARTITION BY a.external_game_id
            ORDER BY a.id
           ) as rn
    FROM artifacts a
    JOIN text_review_details td ON td.artifact_id = a.id
    WHERE a.modality = 'text'
        AND a.scraped_at <= @artifacts_cutoff
        AND (@min_hours_played::int <= 0 OR (td.source_meta->>'hours_played')::int >= @min_hours_played::int)
        AND (coalesce(cardinality(@game_ids::int[]), 0) = 0 OR a.external_game_id = ANY(@game_ids::int[]))
) ranked
WHERE ranked.rn <= @per_game_cap::int
ON CONFLICT (population_id, artifact_id) DO NOTHING;

-- name: FreezeImagePopulation :execrows
INSERT INTO individuals (population_id, artifact_id)
SELECT @population_id::int, ranked.artifact_id
FROM (
    SELECT a.id as artifact_id,
           row_number() OVER (
            PARTITION BY a.external_game_id
            ORDER BY a.id
           ) as rn
    FROM artifacts a
    JOIN image_details img ON img.artifact_id = a.id
    WHERE a.modality = 'image'
        AND a.scraped_at <= @artifacts_cutoff
) ranked
WHERE ranked.rn <= @per_game_cap::int
ON CONFLICT (population_id, artifact_id) DO NOTHING;

-- name: FreezeMultimodalPopulation :execrows
-- a multimodal artifact has both channels, so this joins both detail tables: only artifacts with a body
-- and an image are frozen.
INSERT INTO individuals (population_id, artifact_id)
SELECT @population_id::int, ranked.artifact_id
FROM (
    SELECT a.id as artifact_id,
           row_number() OVER (
            PARTITION BY a.external_game_id
            ORDER BY a.id
           ) as rn
    FROM artifacts a
    JOIN text_review_details td ON td.artifact_id = a.id
    JOIN image_details img ON img.artifact_id = a.id
    WHERE a.modality = 'multimodal'
        AND a.scraped_at <= @artifacts_cutoff
) ranked
WHERE ranked.rn <= @per_game_cap::int
ON CONFLICT (population_id, artifact_id) DO NOTHING;