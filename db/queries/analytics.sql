-- name: RunPatternDistribution :many
-- per game and meso code, the number of reviews in the run where a majority of the completed panel
-- flagged the pattern. drives a run's detection profile: most-flagged patterns and family mix, per game
-- and (summed client-side) across all games. keyed by code, stable across taxonomy versions.
WITH total AS (
    SELECT an.individual_id, count(*)::int AS n
    FROM annotations an
    WHERE an.run_id = sqlc.arg(run_id) AND an.status = 'completed'
    GROUP BY an.individual_id
),
flagged AS (
    SELECT an.individual_id, m.code
    FROM annotations an
    JOIN annotation_patterns ap ON ap.annotation_id = an.id
    JOIN taxonomy_meso_levels m ON m.id = ap.pattern_id
    JOIN total t ON t.individual_id = an.individual_id
    WHERE an.run_id = sqlc.arg(run_id) AND an.status = 'completed'
    GROUP BY an.individual_id, m.code, t.n
    HAVING count(*) * 2 > t.n
)
SELECT a.external_game_id, gd.name AS game_name, f.code, count(*)::int AS reviews
FROM flagged f
JOIN individuals i ON i.id = f.individual_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN game_display gd ON gd.external_game_id = a.external_game_id
GROUP BY a.external_game_id, gd.name, f.code
ORDER BY a.external_game_id, reviews DESC;
