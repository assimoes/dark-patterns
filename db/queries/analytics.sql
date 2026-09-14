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


-- name: AnalysisAnnotations :many
-- Long format for one run: one row per (review x panel model) x pattern with present being a flag
-- that indicates the model detected that pattern in the review
SELECT
    an.individual_id,
    art.external_game_id,
    ra.model_slug,
    m.code,
    (ap.pattern_id IS NOT NULL)::bool AS present
FROM annotations an
JOIN runs r ON r.id = an.run_id
JOIN run_annotators ra ON ra.run_id = an.run_id AND ra.annotator_id = an.annotator_id
JOIN individuals i on i.id = an.individual_id
JOIN artifacts art on art.id = i.artifact_id
CROSS JOIN taxonomy_meso_levels m 
LEFT JOIN annotation_patterns ap ON ap.annotation_id = an.id AND ap.pattern_id = m.id
WHERE an.run_id = sqlc.arg(run_id)
    AND an.status = 'completed'
    AND m.version = r.taxonomy_version
ORDER BY an.individual_id, ra.model_slug, m.code;


-- name: AnalysisGold :many
-- The adjudicated gold for a gold run, long format per (review, pattern, pass). final_label is the
-- human annotator call; direction records whether it confirmed or replaced the panel majority at decision time.
SELECT adj.individual_id,
    m.code,
    adj.pass,
    adj.final_label,
    adj.direction
FROM adjudications adj
JOIN taxonomy_meso_levels m ON m.id = adj.pattern_id
WHERE adj.run_id = sqlc.arg(gold_run_id)
ORDER BY adj.individual_id, m.code, adj.pass;

-- name: AnalysisStatus :many
-- The raw panel coverage for a run
SELECT
    an.individual_id,
    ra.model_slug,
    an.status,
    COALESCE(an.response_meta->>'finish_reason', '')::text AS finish_reason
FROM annotations an
JOIN run_annotators ra on ra.run_id = an.run_id and ra.annotator_id = an.annotator_id
WHERE an.run_id = sqlc.arg(run_id)
ORDER BY an.individual_id, ra.model_slug;

-- name: AnalysisFailures :many
-- Every annotation that did not complete, for one run, with its model, finish reason, token count, and the
-- raw text the model returned that failed to parse -- the input for a per-model failure-mode analysis.
-- raw_response is stored as a JSON string, so #>> '{}' returns the unescaped text the model actually sent.
SELECT
    an.individual_id,
    ra.model_slug,
    an.status,
    COALESCE(an.response_meta->>'finish_reason', '')::text AS finish_reason,
    COALESCE(an.response_meta->>'completion_tokens', '')::text AS completion_tokens,
    COALESCE(an.raw_response #>> '{}', '')::text AS raw_response
FROM annotations an
JOIN run_annotators ra on ra.run_id = an.run_id and ra.annotator_id = an.annotator_id
WHERE an.run_id = sqlc.arg(run_id)
    AND an.status <> 'completed'
ORDER BY ra.model_slug, an.individual_id;

-- name: AnalysisSampleStrata :many
-- The stratified adjudication sample behind a gold run: one row per (sample, review) with the stratum it
-- was drawn from (silent / flagged_split / flagged_majority, classified by the sample's panel-run votes) and
-- the inverse-probability selection weight. Lets the notebook show that the gold is a stratified subset, not
-- the corpus, so the calibration metrics are read as sample-conditional.
SELECT s.id AS sample_id,
    s.panel_run_id,
    si.individual_id,
    si.stratum,
    si.selection_prob
FROM adjudication_samples s
JOIN adjudication_sample_items si ON si.sample_id = s.id
WHERE s.gold_run_id = sqlc.arg(gold_run_id)
ORDER BY s.id, si.individual_id;

-- name: AnalysisSeedVotes :many
-- The per-model panel votes frozen in each adjudication's seed (the panel that was on screen at decision
-- time), lateral-expanded from panel_seed_at_adjudication->'votes'. Lets the notebook prove that the live
-- annotations of the adjudicated run match the frozen seed, so deriving the majority is faithful.
SELECT a.individual_id,
    m.code,
    (v->>'model_slug')::text AS model_slug,
    (v->>'present')::boolean AS present
FROM adjudications a
JOIN taxonomy_meso_levels m ON m.id = a.pattern_id
CROSS JOIN LATERAL jsonb_array_elements(a.panel_seed_at_adjudication->'votes') AS v
WHERE a.run_id = sqlc.arg(gold_run_id)
ORDER BY a.individual_id, m.code;


