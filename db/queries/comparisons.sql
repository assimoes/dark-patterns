-- name: InsertComparison :one
INSERT INTO comparisons (label, run_a_id, run_b_id, contract)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListComparisons :many
SELECT * FROM comparisons ORDER BY created_at DESC;

-- name: GetComparison :one
SELECT * FROM comparisons WHERE id = $1;

-- name: DeleteComparison :exec
DELETE FROM comparisons WHERE id = $1;

-- name: ComparisonReviewDivergence :many
-- the worklist: every review both runs annotated, with how many patterns the two panels disagreed on.
-- this works for any two runs (same or different panels): each panel's verdict on a pattern is whether a
-- strict majority of its completed members flagged it; a flip is a pattern one panel's majority flagged
-- and the other's did not. patterns are keyed by meso code (stable across taxonomy versions). the
-- per-model detail lives in the drill-down.
WITH shared_reviews AS (
    SELECT an.individual_id FROM annotations an WHERE an.run_id = sqlc.arg(run_a) AND an.status = 'completed'
    INTERSECT
    SELECT an.individual_id FROM annotations an WHERE an.run_id = sqlc.arg(run_b) AND an.status = 'completed'
),
a_total AS (
    SELECT an.individual_id, count(*)::int AS n
    FROM annotations an WHERE an.run_id = sqlc.arg(run_a) AND an.status = 'completed'
    GROUP BY an.individual_id
),
b_total AS (
    SELECT an.individual_id, count(*)::int AS n
    FROM annotations an WHERE an.run_id = sqlc.arg(run_b) AND an.status = 'completed'
    GROUP BY an.individual_id
),
a_flagged AS (
    SELECT an.individual_id, m.code
    FROM annotations an
    JOIN annotation_patterns ap ON ap.annotation_id = an.id
    JOIN taxonomy_meso_levels m ON m.id = ap.pattern_id
    JOIN a_total t ON t.individual_id = an.individual_id
    WHERE an.run_id = sqlc.arg(run_a) AND an.status = 'completed'
    GROUP BY an.individual_id, m.code, t.n
    HAVING count(*) * 2 > t.n
),
b_flagged AS (
    SELECT an.individual_id, m.code
    FROM annotations an
    JOIN annotation_patterns ap ON ap.annotation_id = an.id
    JOIN taxonomy_meso_levels m ON m.id = ap.pattern_id
    JOIN b_total t ON t.individual_id = an.individual_id
    WHERE an.run_id = sqlc.arg(run_b) AND an.status = 'completed'
    GROUP BY an.individual_id, m.code, t.n
    HAVING count(*) * 2 > t.n
),
flips AS (
    SELECT COALESCE(a.individual_id, b.individual_id) AS individual_id, count(*)::int AS flips
    FROM a_flagged a
    FULL OUTER JOIN b_flagged b ON a.individual_id = b.individual_id AND a.code = b.code
    WHERE a.code IS NULL OR b.code IS NULL
    GROUP BY COALESCE(a.individual_id, b.individual_id)
)
SELECT sr.individual_id,
    art.external_game_id,
    COALESCE(f.flips, 0)::int AS flips
FROM shared_reviews sr
JOIN individuals i ON i.id = sr.individual_id
JOIN artifacts art ON art.id = i.artifact_id
LEFT JOIN flips f ON f.individual_id = sr.individual_id
ORDER BY COALESCE(f.flips, 0) DESC, art.external_game_id, sr.individual_id;

-- name: ComparisonModelAgreement :many
-- per shared model (in both panels), how many flagged patterns both runs agreed on vs flipped, across
-- the shared reviews. the universe is the union of present-cells for that model; mutual absence is not
-- counted. keyed by the meso code so it is stable across taxonomy versions.
WITH shared_models AS (
    SELECT ra.model_slug FROM run_annotators ra WHERE ra.run_id = sqlc.arg(run_a)
    INTERSECT
    SELECT ra.model_slug FROM run_annotators ra WHERE ra.run_id = sqlc.arg(run_b)
),
shared_reviews AS (
    SELECT an.individual_id FROM annotations an WHERE an.run_id = sqlc.arg(run_a) AND an.status = 'completed'
    INTERSECT
    SELECT an.individual_id FROM annotations an WHERE an.run_id = sqlc.arg(run_b) AND an.status = 'completed'
),
a_cells AS (
    SELECT ra.model_slug, an.individual_id, m.code
    FROM annotations an
    JOIN run_annotators ra ON ra.run_id = an.run_id AND ra.annotator_id = an.annotator_id
    JOIN annotation_patterns ap ON ap.annotation_id = an.id
    JOIN taxonomy_meso_levels m ON m.id = ap.pattern_id
    WHERE an.run_id = sqlc.arg(run_a) AND an.status = 'completed'
        AND ra.model_slug IN (SELECT model_slug FROM shared_models)
        AND an.individual_id IN (SELECT individual_id FROM shared_reviews)
),
b_cells AS (
    SELECT ra.model_slug, an.individual_id, m.code
    FROM annotations an
    JOIN run_annotators ra ON ra.run_id = an.run_id AND ra.annotator_id = an.annotator_id
    JOIN annotation_patterns ap ON ap.annotation_id = an.id
    JOIN taxonomy_meso_levels m ON m.id = ap.pattern_id
    WHERE an.run_id = sqlc.arg(run_b) AND an.status = 'completed'
        AND ra.model_slug IN (SELECT model_slug FROM shared_models)
        AND an.individual_id IN (SELECT individual_id FROM shared_reviews)
),
joined AS (
    SELECT COALESCE(a.model_slug, b.model_slug) AS model_slug,
        (a.code IS NOT NULL) AS in_a,
        (b.code IS NOT NULL) AS in_b
    FROM a_cells a
    FULL OUTER JOIN b_cells b
        ON a.model_slug = b.model_slug
        AND a.individual_id = b.individual_id
        AND a.code = b.code
)
SELECT model_slug,
    count(*) FILTER (WHERE in_a AND in_b)::int AS agreed_present,
    count(*) FILTER (WHERE in_a <> in_b)::int AS flips
FROM joined
GROUP BY model_slug
ORDER BY model_slug;
