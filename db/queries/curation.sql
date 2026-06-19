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
-- Representative selection with common filters and the cutoff
-- Criteria need dynamic SQL
SELECT
FROM artifacts a
JOIN text_review_details td ON td.artifact_id = a.id
WHERE a.modality = 'text'
    AND a.scraped_at <= $1
    AND td.hours_played >= $2;

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
        AND td.hours_played >= @min_hours_played::int
        -- an empty or null game list means all games, otherwise keep only the picked ones. coalesce so
        -- a null array reads as 0, same as empty, instead of dropping every row.
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