-- name: ListGameDisplays :many
-- The curated games with their presentation metadata. external_game_id is the stable id
-- the frontend uses as `gameId`; the rest are display-only fields the pipeline never needed.
SELECT external_game_id, name, short, monetization, display_color
FROM game_display
ORDER BY external_game_id;
 
-- name: CountIndividualsPerGame :many
-- Curated individuals (reviews in a population) per game, across every population.
SELECT a.external_game_id, count(*)::int AS individuals
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
GROUP BY a.external_game_id
ORDER BY a.external_game_id;
 
-- name: ListRunsForDashboard :many
-- One row per run with its population's modality as a human label and its creation time.
-- annotator_ids is the panel the run pinned; members are resolved separately per run.
SELECT
    r.id,
    r.run_type,
    p.modality AS population,
    r.created_at,
    r.annotator_ids
FROM runs r
JOIN populations p ON p.id = r.population_id
ORDER BY r.created_at DESC;
 
-- name: ListMembersForRun :many
-- The panel members of a run: every annotator referenced by runs.annotator_ids, with the kind
-- (llm | human) and a display label. LLM members carry their model name, humans their own label.
SELECT
    an.id,
    an.kind,
    COALESCE(m.name, an.label) AS label
FROM runs r
JOIN annotators an ON an.id = ANY(r.annotator_ids)
LEFT JOIN models m ON m.id = an.model_id
WHERE r.id = sqlc.arg(run_id)
ORDER BY an.kind, an.id;
 
-- name: ReviewStatsPerGame :many
-- Per game: how many reviews are in scope (curated individuals) and how many have at least one
-- completed annotation in any run. The annotated count is distinct individuals, not annotations.
SELECT
    a.external_game_id,
    count(DISTINCT i.id)::int AS reviews,
    count(DISTINCT i.id) FILTER (
        WHERE EXISTS (
            SELECT 1 FROM annotations an
            WHERE an.individual_id = i.id AND an.status = 'completed'
        )
    )::int AS annotated
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
GROUP BY a.external_game_id
ORDER BY a.external_game_id;
 
-- name: ModelStatsForGame :many
-- Per LLM model, the number of completed annotations produced on reviews of one game.
-- annotators -> models gives the model name; only completed annotations are counted.
SELECT
    m.name AS model,
    count(*)::int AS annotated
FROM annotations an
JOIN individuals i ON i.id = an.individual_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN annotators ann ON ann.id = an.annotator_id
JOIN models m ON m.id = ann.model_id
WHERE a.external_game_id = sqlc.arg(external_game_id)
    AND an.status = 'completed'
    AND ann.kind = 'llm'
GROUP BY m.name
ORDER BY annotated DESC, m.name;
 
-- name: ListRunReviews :many
-- The adjudication queue for a run: every distinct individual (review) annotated in this run,
-- with the game it belongs to and the review text/vote/language to render. One row per review.
SELECT DISTINCT
    i.id AS individual_id,
    a.external_game_id,
    td.voted_up,
    td.lang,
    td.body
FROM annotations an
JOIN individuals i ON i.id = an.individual_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN text_review_details td ON td.artifact_id = a.id
WHERE an.run_id = sqlc.arg(run_id)
ORDER BY a.external_game_id, i.id;
 
-- name: ListPresentPatternsForReview :many
-- The patterns the panel marked present on one review in one run: a pattern is present when a
-- majority of the completing raters flagged it. Returns the meso code and name for each.
SELECT
    t.code,
    t.name
FROM taxonomy_meso_levels t
JOIN annotation_patterns ap ON ap.pattern_id = t.id
JOIN annotations an ON an.id = ap.annotation_id
WHERE an.run_id = sqlc.arg(run_id)
    AND an.individual_id = sqlc.arg(individual_id)
    AND an.status = 'completed'
GROUP BY t.id, t.code, t.name
HAVING count(*) * 2 > (
    SELECT count(*) FROM annotations a2
    WHERE a2.run_id = sqlc.arg(run_id)
        AND a2.individual_id = sqlc.arg(individual_id)
        AND a2.status = 'completed'
)
ORDER BY t.code;
 
-- name: GetPatternIDByCode :one
-- Resolve a meso pattern code (e.g. 'PM-1') the frontend sends to its row id for adjudication.
SELECT id FROM taxonomy_meso_levels WHERE code = sqlc.arg(code);
 
-- name: GetGoldRunForReview :one
-- The gold run that adjudications for this review's population are written to. There is one gold
-- run per population; pick the most recent so revisits land on the same row the queue was built from.
SELECT r.id
FROM runs r
JOIN individuals i ON i.population_id = r.population_id
WHERE i.id = sqlc.arg(individual_id)
    AND r.run_type = 'gold'
ORDER BY r.created_at DESC
LIMIT 1;
 
-- name: GetPanelRunForReview :one
-- The llm panel run whose votes seed a decision on this review. One panel per population; pick the
-- most recent so the frozen seed reflects the panel the auditor is actually looking at.
SELECT r.id
FROM runs r
JOIN individuals i ON i.population_id = r.population_id
WHERE i.id = sqlc.arg(individual_id)
    AND r.run_type = 'llm_panel'
ORDER BY r.created_at DESC
LIMIT 1;

-- name: ListPopulationsForGame :many
-- The populations that contain this game's reviews, each with the game's slice: how many of the
-- game's individuals fall in the population, and how many of those have a completed annotation. A
-- population is multi-game (stratified, per-game capped), so this is THIS game's part of it. Counts
-- are per population, never summed across them.
SELECT
    p.id AS population_id,
    p.description,
    p.created_at,
    count(DISTINCT i.id)::int AS reviews,
    count(DISTINCT i.id) FILTER (
        WHERE EXISTS (
            SELECT 1 FROM annotations an
            WHERE an.individual_id = i.id AND an.status = 'completed'
        )
    )::int AS annotated
FROM populations p
JOIN individuals i ON i.population_id = p.id
JOIN artifacts a ON a.id = i.artifact_id
WHERE a.external_game_id = sqlc.arg(external_game_id)
GROUP BY p.id, p.description, p.created_at
ORDER BY p.id;
 
-- name: ModelStatsForGamePopulation :many
-- Per LLM model, the number of DISTINCT reviews of one game annotated within one population. Counting
-- distinct individuals (not annotation rows) and scoping to a single population makes the models
-- comparable: inside one population they all share the same work set, so a complete run shows every
-- model at the population's slice size, not a runaway sum across runs.
SELECT
    m.name AS model,
    count(DISTINCT i.id)::int AS annotated
FROM annotations an
JOIN individuals i ON i.id = an.individual_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN annotators ann ON ann.id = an.annotator_id
JOIN models m ON m.id = ann.model_id
WHERE a.external_game_id = sqlc.arg(external_game_id)
    AND i.population_id = sqlc.arg(population_id)
    AND an.status = 'completed'
    AND ann.kind = 'llm'
GROUP BY m.name
ORDER BY annotated DESC, m.name;
 
-- name: PanelForPopulation :many
-- The panel that worked a population: every annotator frozen onto any of the population's runs, with
-- its kind (llm | human) and a display label (the model name for an llm, the annotator's own label
-- for a human). run_annotators is the single source of "who annotates this run".
SELECT DISTINCT
    an.kind,
    COALESCE(m.name, an.label) AS label
FROM run_annotators ra
JOIN runs r ON r.id = ra.run_id
JOIN annotators an ON an.id = ra.annotator_id
LEFT JOIN models m ON m.id = an.model_id
WHERE r.population_id = sqlc.arg(population_id)
ORDER BY an.kind, label;