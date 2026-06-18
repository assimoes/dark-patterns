-- name: SampleStratifiedIndividuals :many
-- Fixed N individuals per game from the population, ordered deterministically
-- Always yields the same subset.
-- Drops any review where a panel member didn't 'complete' (>=1 parse_error) for the panel run
SELECT i.id AS individual_id, a.external_game_id
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
WHERE i.population_id = sqlc.arg(population_id)
  AND NOT EXISTS (                                  -- ignore reviews with any non-completed annotation
      SELECT 1 FROM annotations an
      WHERE an.individual_id = i.id
        AND an.run_id = sqlc.arg(panel_run_id)
        AND an.status <> 'completed'
  )
  AND (
      SELECT count(*)::int                          -- this review's rank within its game
      FROM individuals i2
      JOIN artifacts a2 ON a2.id = i2.artifact_id
      WHERE i2.population_id = i.population_id
        AND a2.external_game_id = a.external_game_id
        AND i2.id <= i.id
        AND NOT EXISTS (
            SELECT 1 FROM annotations an2
            WHERE an2.individual_id = i2.id
              AND an2.run_id = sqlc.arg(panel_run_id)
              AND an2.status <> 'completed'
        )
  ) <= sqlc.arg(per_game)
ORDER BY a.external_game_id, i.id;


-- name: GetReviewText :one
-- The review body to render for the auditor
SELECT td.body
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
JOIN text_review_details td ON td.artifact_id = a.id
WHERE i.id = sqlc.arg(individual_id);

-- name: GetPanelVoteForCell :one
-- The panel's verdict for one (run, individual, pattern)
SELECT
    count(*) FILTER (WHERE ap.pattern_id IS NOT NULL) as n_present,
    count(*) as n_total
FROM annotations an
LEFT JOIN annotation_patterns ap
    ON ap.annotation_id = an.id AND ap.pattern_id = sqlc.arg(pattern_id)
WHERE an.run_id = sqlc.arg(panel_run_id)
    AND an.individual_id = sqlc.arg(individual_id)
    AND an.status = 'completed';

-- name: ListPanelVotesForCell :many
-- Per-rater verdicts with each model's own evidence and explanation
SELECT an.annotator_id,
    ra.model_slug,
    (ap.pattern_id IS NOT NULL)::bool as present,
    COALESCE(ap.evidence, '')::text as evidence,
    COALESCE(ap.explanation, '')::text as explanation
FROM annotations an
JOIN run_annotators ra ON ra.run_id = an.run_id AND ra.annotator_id = an.annotator_id
LEFT JOIN annotation_patterns ap
    ON ap.annotation_id = an.id AND ap.pattern_id = sqlc.arg(pattern_id)
WHERE an.run_id = sqlc.arg(panel_run_id)
    AND an.individual_id = sqlc.arg(individual_id)
    AND an.status = 'completed'
ORDER BY an.annotator_id;

-- name: UpsertAdjudication :exec
-- Re-saving a review updates the decision (the auditor is deliberately changing it) and re-freezes
-- the panel seed at the new decision time. pass keeps the open and blind labels for a cell apart, so
-- saving one never touches the other.
INSERT INTO adjudications (
    run_id, individual_id, pattern_id, final_label, direction, adjudicator_id, panel_seed_at_adjudication, pass
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (run_id, individual_id, pattern_id, pass) DO UPDATE
SET final_label = excluded.final_label,
    direction = excluded.direction,
    adjudicator_id = excluded.adjudicator_id,
    panel_seed_at_adjudication = excluded.panel_seed_at_adjudication,
    decided_at = now();

-- name: ListDedicedCells :many
-- gold run
SELECT individual_id, pattern_id FROM adjudications WHERE run_id = sqlc.arg(run_id);

-- name: GetMesoPatternCodes :many
SELECT id, code, name FROM taxonomy_meso_levels WHERE version = sqlc.arg(version) ORDER BY code;

-- name: ListTaxonomy :many
-- The full MESO codebook for a version: code, name, definition, and the family it sits under.
SELECT m.id, m.code, m.name, m.description, h.name AS family
FROM taxonomy_meso_levels m
JOIN taxonomy_high_levels h ON h.id = m.parent_id
WHERE m.version = sqlc.arg(version)
ORDER BY m.id;

-- name: ListReviewDetections :many
-- Every panel member's detection for one review, across all patterns: who flagged what, with their
-- evidence and explanation. annotation_patterns holds positives only, so a row means that model
-- detected that pattern on this review.
SELECT ap.pattern_id,
    ra.model_slug,
    COALESCE(ap.evidence, '')::text as evidence,
    COALESCE(ap.explanation, '')::text as explanation
FROM annotations an
JOIN run_annotators ra ON ra.run_id = an.run_id AND ra.annotator_id = an.annotator_id
JOIN annotation_patterns ap ON ap.annotation_id = an.id
WHERE an.run_id = sqlc.arg(panel_run_id)
    AND an.individual_id = sqlc.arg(individual_id)
    AND an.status = 'completed'
ORDER BY ap.pattern_id, ra.model_slug;

-- name: CountCompletedRaters :one
-- How many panel members completed this review (the denominator for every pattern's vote).
SELECT count(*)::int as n_total
FROM annotations
WHERE run_id = sqlc.arg(panel_run_id)
    AND individual_id = sqlc.arg(individual_id)
    AND status = 'completed';

-- name: ListReviewAdjudications :many
-- Existing gold labels for one review in one pass, to pre-fill the checkboxes on revisit. the blind
-- read asks for pass='blind' so it never sees the open-pass labels, and the other way round.
SELECT pattern_id, final_label
FROM adjudications
WHERE run_id = sqlc.arg(run_id) AND individual_id = sqlc.arg(individual_id) AND pass = sqlc.arg(pass);

-- name: CountAdjudicationsPerReview :many
-- gold run: how many patterns are decided per review, to show progress on the worklist.
SELECT individual_id, count(*)::int as decided
FROM adjudications
WHERE run_id = sqlc.arg(run_id)
GROUP BY individual_id;
